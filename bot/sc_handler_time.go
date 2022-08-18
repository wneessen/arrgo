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
			Type:        discordgo.EmbedTypeArticle,
			Title:       "It's time, Matey!",
			Description: fmt.Sprintf("The current bot time is: <t:%d>", time.Now().Unix()),
		},
	}
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &e}); err != nil {
		return err
	}
	return nil
}
