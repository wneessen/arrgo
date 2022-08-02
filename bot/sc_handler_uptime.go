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
	r := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("I was started on: %s, so I am running for %s now",
				b.StartTimeString(), td.String()),
		},
	}
	if err := s.InteractionRespond(i.Interaction, &r); err != nil {
		return fmt.Errorf("failed to respond to /uptime request: %w", err)
	}
	return nil
}
