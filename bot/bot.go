package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
	"github.com/wneessen/arrgo/config"
	"math/rand"
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
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates |
		discordgo.IntentsDirectMessages | discordgo.IntentsGuildPresences

	// Add handlers
	b.Session.AddHandlerOnce(b.ReadyHandler)
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

// StartTime returns the time when the bot was last initalized
func (b *Bot) StartTime() string {
	return b.st.Format(time.RFC1123)
}

// ReadyHandler updates the Bot's session data
func (b *Bot) ReadyHandler(s *discordgo.Session, ev *discordgo.Ready) {
	ll := b.Log.With().Str("context", "bot.ReadyHandler").Str("sessionID", ev.SessionID).Logger()
	ll.Debug().Msg("bot reached the 'ready' state...")

	usd := &discordgo.UpdateStatusData{Status: "online"}
	usd.Activities = make([]*discordgo.Activity, 1)
	usd.Activities[0] = &discordgo.Activity{
		Name: fmt.Sprintf("ArrGo v%s", Version),
		Type: discordgo.ActivityTypeGame,
		URL:  "https://github.com/wneessen/arrgo",
	}

	err := s.UpdateStatusComplex(*usd)
	if err != nil {
		ll.Error().Msgf("failed to set bot's ready state: %s", err)
	}
}

// RegisterSlashCommands will fetch the list of available slash commands and register them with the Guild
// if not present yet
func (b *Bot) RegisterSlashCommands() error {
	ll := b.Log.With().Str("context", "bot.RegisterSlashCommands").Logger()

	// Get a list of currently registered slash commands
	rcl, err := b.Session.ApplicationCommands(b.Session.State.User.ID, "")
	if err != nil {
		return fmt.Errorf("failed to fetch list registered slash commands: %w", err)
	}

	for _, sc := range b.getSlashCommands() {
		n := true
		c := false
		for _, rc := range rcl {
			if sc.Name == rc.Name && sc.Description == rc.Description {
				ll.Debug().Msgf("slash command %s already registered. Skipping.", rc.Name)
				n = false
				c = false
				break
			}
			if sc.Name == rc.Name && sc.Description != rc.Description {
				ll.Debug().Msgf("slash command %s changed. Updating.", rc.Name)
				n = false
				c = true
				break
			}
		}
		if n || c {
			go func(s *discordgo.ApplicationCommand, e bool) {
				rn := rand.Int31n(2000) + 1000
				rd, _ := time.ParseDuration(fmt.Sprintf("%dms", rn))
				ll.Debug().Msgf("[%s] delaying registration/update for %f seconds", s.Name, rd.Seconds())
				time.Sleep(rd)
				if e {
					ll.Debug().Msgf("[%s] updating slash command...", s.Name)
					_, err := b.Session.ApplicationCommandEdit(b.Session.State.User.ID, "", s.ID, s)
					if err != nil {
						ll.Error().Msgf("[%s] failed to update slash command: %s", s.Name, err)
						return
					}
					ll.Debug().Msgf("[%s] slash command successfully updated...", s.Name)
				}
				if !e {
					ll.Debug().Msgf("[%s] registering slash command...", s.Name)
					_, err := b.Session.ApplicationCommandCreate(b.Session.State.User.ID, "", s)
					if err != nil {
						ll.Error().Msgf("[%s] failed to register slash command: %s", s.Name, err)
						return
					}
					ll.Debug().Msgf("[%s] slash command successfully registered...", s.Name)
				}
			}(sc, c)
		}
	}
	return nil
}
