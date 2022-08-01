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

		// version returns the current version information
		{
			Name:        "version",
			Description: "Tells you some information about the bot",
		},
	}
}

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
		"time":    b.SlashCmdTime,
		"version": b.SlashCmdVersion,
	}

	// Check if provided command is available and process it
	if h, ok := sh[i.ApplicationCommandData().Name]; ok {
		if err := h(s, i); err != nil {
			ll.Error().Msgf("failed to process /%s command: %s", i.ApplicationCommandData().Name, err)
		}
	}
}
