package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"time"
)

// SlashCmdUptime handles the /uptime slash command
func (b *Bot) SlashCmdUptime(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ut := time.Now().Unix() - b.StartTimeUnix()
	td, err := time.ParseDuration(fmt.Sprintf("%ds", ut))
	if err != nil {
		return fmt.Errorf("failed to parse time difference: %w", err)
	}
	e := []*discordgo.MessageEmbed{
		{
			Type: discordgo.EmbedTypeRich,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "Forrest Gump would be proud...",
					Value: fmt.Sprintf("I started running: <t:%d> and haven't stopped since... "+
						"which means I've been running for %s now!", b.StartTimeUnix(), td.String()),
					Inline: false,
				},
			},
		},
	}
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: e}); err != nil {
		return err
	}
	return nil
}
