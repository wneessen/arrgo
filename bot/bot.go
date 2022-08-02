package bot

import (
	"crypto/rand"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/wneessen/arrgo/config"
	"github.com/wneessen/arrgo/model"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"
)

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
	b.Model = model.New(db)

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

// randNum returns a random number with a maximum value of n
func (b *Bot) randNum(n int) (int, error) {
	if n <= 0 {
		return 0, fmt.Errorf("provided number is <= 0: %d", n)
	}
	mbi := big.NewInt(int64(n))
	if !mbi.IsUint64() {
		return 0, fmt.Errorf("big.NewInt() generation returned negative value: %d", mbi)
	}
	rn64, err := rand.Int(rand.Reader, mbi)
	if err != nil {
		return 0, err
	}
	rn := int(rn64.Int64())
	if rn < 0 {
		return 0, fmt.Errorf("generated random number does not fit as int64: %d", rn64)
	}
	return rn, nil
}
