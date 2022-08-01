package bot

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

// getSlashCommands returns a list of slash commands that will be registered for the bot
func (b *Bot) getSlashCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		// time returns the current time back to the user
		{
			Name:        "time",
			Description: "Let's you know how late it currently is",
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

	switch strings.ToLower(i.ApplicationCommandData().Name) {
	case "time":
		ll.Debug().Msg("time slash command requested")
		if err := b.SlashCmdTime(s, i); err != nil {
			ll.Error().Msgf("failed to process time request: %s", err)
		}
	default:
		ll.Warn().Msgf("unknown slash command: %s", i.ApplicationCommandData().Name)
	}
}
