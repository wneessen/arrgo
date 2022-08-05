package bot

import (
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

// FHTimer defines the maximum random number for the FH spammer timer (in minutes)
const FHTimer = 2

// Bot represents the bot instance
type Bot struct {
	Log     zerolog.Logger
	Config  *config.Config
	Session *discordgo.Session
	Model   model.Model

	st time.Time
}

// New initalizes a new Bot instance
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
	ll.Debug().Msg("initalizing bot...")

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

	// Times events
	rn := FHTimer
	rn, err = crypto.RandNum(FHTimer)
	if err != nil {
		ll.Warn().Msgf("failed to generate random number for FH timer: %s", err)
		rn = FHTimer
	}
	fht := time.NewTicker(time.Duration(int64(rn)+FHTimer) * time.Minute)
	defer fht.Stop()

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
			if err := b.ScheduledEventSoTFlameheart(); err != nil {
				ll.Error().Msgf("failed to process scheuled flameheart event: %s", err)
			}

			// Reset the duration
			rn, err = crypto.RandNum(FHTimer)
			if err != nil {
				ll.Warn().Msgf("failed to generate random number for FH timer: %s", err)
				rn = FHTimer
			}
			fht.Reset(time.Duration(int64(rn)+FHTimer) * time.Minute)
		}
	}
}

// StartTimeString returns the time when the bot was last initalized
func (b *Bot) StartTimeString() string {
	return b.st.Format(time.RFC1123)
}

// StartTimeUnix returns the time when the bot was last initalized
func (b *Bot) StartTimeUnix() int64 {
	return b.st.Unix()
}
