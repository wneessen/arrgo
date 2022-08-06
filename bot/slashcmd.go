package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/wneessen/arrgo/crypto"
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
			Description: "Configure certain aspects of your ArrBot instance (admin-only)",
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

		// register registers the requesting user with the bot
		{
			Name:        "register",
			Description: "Regsiters your user with ArrGo so you can use certain user-specific features",
		},

		// setrat stores the SoT authentication token in the Bot's database
		{
			Name:        "setrat",
			Description: "Stores the Sea of Thieves authentication token in the Bot's database",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "rat-cookie",
					Description: "Full SoT authentication cookie string as provided by the SoT-RAT-Extractor",
					Required:    true,
				},
			},
		},

		// achievement gets the users latest achievement from the SoT API
		{
			Name:        "achievement",
			Description: "Returns your latest achievement in Sea of Thieves to you",
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
			sc.ApplicationID = rc.ApplicationID
			sc.ID = rc.ID
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
				rn, err := crypto.RandNum(2000)
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
		"time":        b.SlashCmdTime,
		"uptime":      b.SlashCmdUptime,
		"version":     b.SlashCmdVersion,
		"flameheart":  b.SlashCmdSoTFlameheart,
		"config":      b.SlashCmdConfig,
		"register":    b.SlashCmdRegister,
		"setrat":      b.SlashCmdSetRAT,
		"achievement": b.SlashCmdSoTAchievement,
	}

	// Define list of slash commands that should use ephemeral messages
	el := map[string]bool{
		"register": true,
		"setrat":   true,
		"config":   true,
		"version":  true,
	}

	// Check if provided command is available and process it
	if h, ok := sh[i.ApplicationCommandData().Name]; ok {
		r := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: ""},
		}
		if _, ok := el[i.ApplicationCommandData().Name]; ok {
			r.Data.Flags = uint64(discordgo.MessageFlagsEphemeral)
		}
		err := s.InteractionRespond(i.Interaction, r)
		if err != nil {
			ll.Error().Msgf("failed to defer the /%s command request: %s",
				i.ApplicationCommandData().Name, err)
			return
		}
		if err := h(s, i); err != nil {
			ll.Error().Msgf("failed to process /%s command: %s", i.ApplicationCommandData().Name, err)
			e := []*discordgo.MessageEmbed{
				{
					Type: discordgo.EmbedTypeRich,
					Fields: []*discordgo.MessageEmbedField{
						{
							Name: "Something went wrong!",
							Value: fmt.Sprintf("I am sorry, but I was not able to process your request: %s",
								err),
							Inline: false,
						},
					},
				},
			}
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: e})
		}
	}
}
