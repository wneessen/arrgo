package bot

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/wneessen/arrgo/crypto"
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
				{
					Name:        "announce-sot-summary",
					Description: "Enable/Disable posting of SoT play summaries to the system/announce channel",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Name:        "enable",
							Description: "Announce Sea of Thieves play summaries",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
						{
							Name:        "disable",
							Description: "Do not announce Sea of Thieves play summaries",
							Type:        discordgo.ApplicationCommandOptionSubCommand,
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommandGroup,
				},
			},
		},

		// override allows to override some defaults
		{
			Name:        "override",
			Description: "Override some default settings of your ArrBot instance (admin-only)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "announce-channel",
					Description: "Override the default system channel for bot related announcements",
					Required:    true,
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

		// season gets the users season renown progress from the SoT API
		{
			Name:        "season",
			Description: "Returns your renown progress in the current Sea of Thieves season to you",
		},

		// balance gets the users current gold/doubloon/ac balance from the SoT API
		{
			Name:        "balance",
			Description: "Returns your current Sea of Thieves gold/doubloon/ancient coins balance",
		},

		// traderoutes announces the currently active tradring routes (from rarethief.com)
		{
			Name:        "traderoutes",
			Description: "Returns the currently active trade routes in Sea of Thieves",
		},

		// overview get the users statistics overview from the SoT API
		{
			Name:        "overview",
			Description: "Returns an overview of some general stats of your Sea of Thieves pirate",
		},

		// historic compares the current Sea of Thives user stats with history data
		{
			Name:        "compare",
			Description: "Compares the current Sea of Thieves users stats with historic data (in hours)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "hours",
					Description: "The backwards duration in hours that the historic data should be calculated from",
					Required:    true,
				},
			},
		},

		// dailydeeds get the currently active daily deeds in Sea of Thieves
		{
			Name:        "dailydeeds",
			Description: "Returns the currently active Sea of Thieves daily deeds",
		},

		// ledger provides the current leaderboard position in the different emissary ledgers
		{
			Name:        "ledger",
			Description: "Returns your current leaderboard position in the different emissary ledgers",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "emissary-faction",
					Description: "Name of the emissary faction",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Athena's Fortune", Value: "athena"},
						{Name: "Gold Hoarder", Value: "hoarder"},
						{Name: "Merchant Alliance", Value: "merchant"},
						{Name: "Order of Souls", Value: "order"},
						{Name: "Reaper's Bone", Value: "reaper"},
					},
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
			if sc.Name == rc.Name && len(sc.Options) != len(rc.Options) {
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

// RemoveSlashCommands will fetch the list of registered slash commands and remove them
func (b *Bot) RemoveSlashCommands() error {
	ll := b.Log.With().Str("context", "bot.RegisterSlashCommands").Logger()

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

	// Open the websocket and begin listening.
	err = b.Session.Open()
	if err != nil {
		return fmt.Errorf("failed to open websocket to listen: %w", err)
	}

	// Get a list of currently registered slash commands
	rcl, err := b.Session.ApplicationCommands(b.Session.State.User.ID, "")
	if err != nil {
		return fmt.Errorf("failed to fetch list registered slash commands: %w", err)
	}
	for _, rc := range rcl {
		if err := b.Session.ApplicationCommandDelete(rc.ApplicationID, "", rc.ID); err != nil {
			ll.Warn().Msgf("failed to remove %s command: %s", rc.ID, err)
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
		"override":    b.SlashCmdConfig,
		"register":    b.SlashCmdRegister,
		"setrat":      b.SlashCmdSetRAT,
		"achievement": b.SlashCmdSoTAchievement,
		"season":      b.SlashCmdSoTSeasonProgress,
		"balance":     b.SlashCmdSoTBalance,
		"traderoutes": b.SlashCmdSoTTradeRoutes,
		"overview":    b.SlashCmdSoTOverview,
		"compare":     b.SlashCmdSoTCompare,
		"dailydeeds":  b.SlashCmdSoTDailyDeeds,
		"ledger":      b.SlashCmdSoTLedger,
	}

	// Define list of slash commands that should use ephemeral messages
	el := map[string]bool{
		"register": true,
		"setrat":   true,
		"config":   true,
		"override": true,
		"version":  true,
	}

	// Check if provided command is available and process it
	if h, ok := sh[i.ApplicationCommandData().Name]; ok {
		r := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{Content: ""},
		}
		if _, ok := el[i.ApplicationCommandData().Name]; ok {
			r.Data.Flags = discordgo.MessageFlagsEphemeral
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
					Type: discordgo.EmbedTypeArticle,
					Description: fmt.Sprintf("I am sorry, but I was not able to process your request: %s",
						err),
					Title: "Oh no! Something went wrong!",
				},
			}
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &e})
		}
	}
}
