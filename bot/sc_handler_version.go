package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

// SlashCmdVersion handles the /version slash command
func (b *Bot) SlashCmdVersion(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	e := []*discordgo.MessageEmbed{
		{
			Type: discordgo.EmbedTypeRich,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name: "Oh look! It's me!",
					Value: fmt.Sprintf("I am ArrBot (Version v%s)! Your Sea of Thieves themed discord bot. "+
						"Nice to meet you!", Version),
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
