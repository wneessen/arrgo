package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"time"
)

// getSlashCommands returns a list of slash commands that will be registered for the bot
func (b *Bot) getSlashCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		// time returns the current time back to the user
		{
			Name:        "time",
			Description: "Let's you know how late it currently is",
		},

		// uptime returns the current time back to the user
		{
			Name:        "uptime",
			Description: "Let's you know how long the bot has been running",
		},

		// version returns the current version information
		{
			Name:        "version",
			Description: "Tells you some information about the bot",
		},

		// flameheart returns a random SoT Flameheart quote in all caps
		{
			Name:        "flameheart",
			Description: "Returns a random quote from Captain Flameheart",
		},

		// config allows to configure certain settings of the ArrBot
		{
			Name:        "config",
			Description: "Configure certain aspects of your ArrBot instance",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "flameheart-spam",
					Description: "Enable/Disable the random Captain Flameheart quote spam",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "enable",
							Description: "Enable Captain Flameheart",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
						{
							Name:        "disable",
							Description: "Disable Captain Flameheart",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommandGroup,
				},
			},
		},
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
				rn, err := b.randNum(2000)
				if err != nil {
					ll.Error().Msgf("failed to generate random number: %s", err)
					return
				}
				rn += 1000
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

// SlashCommandHandler is the central handler method for all slash commands. It will look up
// the name of the received SC-handler event in a map and when found execute the corresponding
// method
func (b *Bot) SlashCommandHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ll := b.Log.With().Str("context", "bot.SlashCommandHandler").
		Str("command_type", i.Data.Type().String()).
		Str("command_name", i.ApplicationCommandData().Name).Logger()

	// We only process ApplicationCommands
	if i.Data.Type().String() != "ApplicationCommand" {
		return
	}

	// Define list of slash command handler methods
	sh := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error{
		"time":       b.SlashCmdTime,
		"uptime":     b.SlashCmdUptime,
		"version":    b.SlashCmdVersion,
		"flameheart": b.SlashCmdSoTFlameheart,
		"config":     b.SlashCmdConfig,
	}

	// Check if provided command is available and process it
	if h, ok := sh[i.ApplicationCommandData().Name]; ok {
		if err := h(s, i); err != nil {
			ll.Error().Msgf("failed to process /%s command: %s", i.ApplicationCommandData().Name, err)
		}
	}
}
