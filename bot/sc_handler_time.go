package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"time"
)

// SlashCmdTime handles the /time slash command
func (b *Bot) SlashCmdTime(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	e := []*discordgo.MessageEmbed{
		{
			Type: discordgo.EmbedTypeRich,
			Fields: []*discordgo.MessageEmbedField{
				{
					Name:   "It's time, Matey!",
					Value:  fmt.Sprintf("The current bot time is: <t:%d>", time.Now().Unix()),
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
