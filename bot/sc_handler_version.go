package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

// SlashCmdVersion handles the /version slash command
func (b *Bot) SlashCmdVersion(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	r := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("This is ArrBot (Version v%s)! Your Sea of Thieves themed discord bot. "+
				"Nice to meet you!", Version),
		},
	}
	if err := s.InteractionRespond(i.Interaction, &r); err != nil {
		return fmt.Errorf("failed to respond to /time request: %w", err)
	}
	return nil
}
