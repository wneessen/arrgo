package bot

import (
	"github.com/bwmarrin/discordgo"
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
	}
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
	}

	// Check if provided command is available and process it
	if h, ok := sh[i.ApplicationCommandData().Name]; ok {
		if err := h(s, i); err != nil {
			ll.Error().Msgf("failed to process /%s command: %s", i.ApplicationCommandData().Name, err)
		}
	}
}
