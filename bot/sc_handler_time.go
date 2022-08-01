package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"time"
)

// SlashCmdTime handles the /time slash command
func (b *Bot) SlashCmdTime(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	r := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("The current bot time is: <t:%d>", time.Now().Unix()),
		},
	}
	if err := s.InteractionRespond(i.Interaction, &r); err != nil {
		return fmt.Errorf("failed to respond to /time request: %w", err)
	}
	return nil
}
