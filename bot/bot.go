package bot

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/wneessen/arrgo/config"
	"github.com/wneessen/arrgo/crypto"
	"github.com/wneessen/arrgo/model"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// List of Sea of Thieves API endpoints
const (
	ApiURLSoTAchievements = "https://www.seaofthieves.com/api/profilev2/achievements"
	ApiURLSoTSeasons      = "https://www.seaofthieves.com/api/profilev2/seasons-progress"
	ApiURLSoTUserBalance  = "https://www.seaofthieves.com/api/profilev2/balance"
	ApiURLSoTUserOverview = "https://www.seaofthieves.com/api/profilev2/overview"
	ApiURLSoTEventHub     = "https://www.seaofthieves.com/event-hub"
	ApiURLRTTradeRoutes   = "https://maps.seaofthieves.rarethief.com/js/trade_routes.js"
	ApiURLSoTLedger       = "https://www.seaofthieves.com/api/ledger/friends"
	AssetsBaseURL         = "https://github.com/wneessen/arrgo/raw/main/assets"
)

const (
	ErrFailedHTTPClient          = "failed to generate new HTTP client: %w"
	ErrFailedRetrieveUserStatsDB = "failed retrieve user status from DB: %s"
	ErrFailedGuildLookupDB       = "failed to look up guild in database: %w"
)

// Bot represents the bot instance
type Bot struct {
	Log     zerolog.Logger
	Config  *config.Config
	Session *discordgo.Session
	Model   model.Model

	st time.Time
}

// New initializes a new Bot instance
func New(l zerolog.Logger, c *config.Config) (*Bot, error) {
	b := &Bot{
		Config: c,
		Log:    l,
		st:     time.Now(),
	}
	if c.Discord.Token == "" {
		if t := os.Getenv("ARRGO_TOKEN"); t == "" {
			return nil, fmt.Errorf("no discord token found in config file %q or environment", c.ConfFilePath())
		} else {
			c.Discord.Token = t
		}
	}

	// Connect to DB model
	db, err := b.OpenDB(c)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	b.Model = model.New(db, c)

	// We require a global encryption key
	if c.Data.EncryptionKey == "" || len(c.Data.EncryptionKey) != config.CryptoKeyLen {
		b.Log.Warn().Msgf("no/invalid encryption key in configuration file... generating key...")
		cs, err := crypto.RandomStringSecure(config.CryptoKeyLen, true, false)
		if err != nil {
			b.Log.Error().Msgf("failed to generate encryption key: %s", err)
			os.Exit(1)
		}
		b.Log.Info().Msg("encryption key generated... please add the following key to your config...")
		b.Log.Info().Msgf(`enc_key = "%s"`, cs)
		os.Exit(0)
	}

	return b, nil
}

// Run executes the Bot's main loop
func (b *Bot) Run() error {
	ll := b.Log.With().Str("context", "bot.Run").Logger()
	ll.Debug().Msg("initializing bot...")

	dg, err := discordgo.New("Bot " + b.Config.Discord.Token)
	if err != nil {
		return fmt.Errorf("failed to create discord session: %w", err)
	}
	b.Session = dg

	// Define list of events we want to see
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildVoiceStates | discordgo.IntentsDirectMessages |
		discordgo.IntentsGuildPresences | discordgo.IntentsMessageContent |
		discordgo.IntentsGuildIntegrations

	// Add handlers
	b.Session.AddHandlerOnce(b.ReadyHandler)
	b.Session.AddHandler(b.GuildCreate)
	b.Session.AddHandler(b.GuildDelete)
	b.Session.AddHandler(b.SlashCommandHandler)
	b.Session.AddHandler(b.UserPlaySoT)

	// Open the websocket and begin listening.
	err = b.Session.Open()
	if err != nil {
		return fmt.Errorf("failed to open websocket to listen: %w", err)
	}

	// Register/Update slash commands
	if err := b.RegisterSlashCommands(); err != nil {
		ll.Error().Msgf("slash command registration failed: %s", err)
	}

	// We need a signal channel
	sc := make(chan os.Signal, 1)
	signal.Notify(sc)

	// Timer events
	rd, err := crypto.RandDuration(b.Config.Timer.FHSpam, "m")
	if err != nil {
		ll.Warn().Msgf("failed to generate random number for FH timer: %s", err)
		rd = time.Minute * time.Duration(b.Config.Timer.FHSpam)
	}
	fht := time.NewTicker(rd)
	defer fht.Stop()
	trt := time.NewTicker(b.Config.Timer.TRUpdate)
	defer trt.Stop()
	ust := time.NewTicker(b.Config.Timer.USUpdate)
	defer ust.Stop()
	rct := time.NewTicker(b.Config.Timer.RCCheck)
	defer rct.Stop()
	ddt := time.NewTicker(b.Config.Timer.DDUpdate)
	defer ddt.Stop()

	// Perform an update for all scheduled update tasks once if first-run flag is set
	if b.Config.GetFirstRun() {
		go func() {
			if err := b.ScheduledEventUpdateTradeRoutes(); err != nil {
				b.Log.Error().Msgf("failed to update trade routes: %s", err)
			}
			if err := b.ScheduledEventUpdateUserStats(); err != nil {
				b.Log.Error().Msgf("failed to update user stats: %s", err)
			}
			if err := b.ScheduledEventUpdateDailyDeeds(); err != nil {
				b.Log.Error().Msgf("failed to update daily deeds: %s", err)
			}
		}()
	}

	// Wait here until CTRL-C or other term signal is received.
	ll.Info().Msg("bot successfully initialized and connected. Press CTRL-C to exit.")
	for {
		select {
		case rs := <-sc:
			if rs == syscall.SIGKILL ||
				rs == syscall.SIGABRT ||
				rs == syscall.SIGINT ||
				rs == syscall.SIGTERM {
				ll.Warn().Msgf("received %s signal. Exiting.", rs)

				// Cleanly close down the Discord session.
				if err := b.Session.Close(); err != nil {
					ll.Error().Msgf("failed to gracefully close discord session: %s", err)
				}
				return nil
			}
		case <-fht.C:
			go func() {
				if err := b.ScheduledEventSoTFlameheart(); err != nil {
					ll.Error().Msgf("failed to process scheuled flameheart event: %s", err)
				}
			}()

			// Reset the duration
			rd, err := crypto.RandDuration(b.Config.Timer.FHSpam, "m")
			if err != nil {
				ll.Warn().Msgf("failed to generate random number for FH timer: %s", err)
				rd = time.Minute * time.Duration(b.Config.Timer.FHSpam)
			}
			fht.Reset(rd)
		case <-trt.C:
			go func() {
				if err := b.ScheduledEventUpdateTradeRoutes(); err != nil {
					ll.Error().Msgf("failed to process scheuled traderoute update event: %s", err)
				}
			}()
		case <-ust.C:
			go func() {
				if err := b.ScheduledEventUpdateUserStats(); err != nil {
					ll.Error().Msgf("failed to process scheuled traderoute update event: %s", err)
				}
			}()
		case <-rct.C:
			go func() {
				if err := b.ScheduledEventCheckRATCookies(); err != nil {
					ll.Error().Msgf("failed to process scheuled RAT cookie check event: %s", err)
				}
			}()
		case <-ddt.C:
			go func() {
				if err := b.ScheduledEventUpdateDailyDeeds(); err != nil {
					ll.Error().Msgf("failed to process scheuled daily deeds update event: %s", err)
				}
			}()
		}
	}
}

// StartTimeString returns the time when the bot was last initialized
func (b *Bot) StartTimeString() string {
	return b.st.Format(time.RFC1123)
}

// StartTimeUnix returns the time when the bot was last initialized
func (b *Bot) StartTimeUnix() int64 {
	return b.st.Unix()
}

// NewRequester returns a Requester based on if it's a channel interaction or DM
func (b *Bot) NewRequester(i *discordgo.Interaction) (*Requester, error) {
	r := &Requester{nil, b.Model.User, nil}
	if i.User != nil {
		u, err := b.Model.User.GetByUserID(i.User.ID)
		if err != nil {
			switch {
			case errors.Is(err, model.ErrUserNotExistent):
				return r, fmt.Errorf("user is not registered")
			default:
				return r, err
			}
		}
		r.User = u
	}
	if i.Member != nil {
		r.Member = i.Member
	}
	return r, nil
}
