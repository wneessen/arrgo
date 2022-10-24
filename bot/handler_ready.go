package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// ReadyHandler updates the Bot's session data
func (b *Bot) ReadyHandler(s *discordgo.Session, ev *discordgo.Ready) {
	ll := b.Log.With().Str("context", "bot.ReadyHandler").Str("sessionID", ev.SessionID).Logger()
	ll.Debug().Msg("bot reached the 'ready' state...")

	usd := &discordgo.UpdateStatusData{Status: "online"}
	usd.Activities = make([]*discordgo.Activity, 1)
	usd.Activities[0] = &discordgo.Activity{
		Name: fmt.Sprintf("ArrGo v%s", Version),
		Type: discordgo.ActivityTypeGame,
		URL:  "https://github.com/wneessen/arrgo",
	}

	err := s.UpdateStatusComplex(*usd)
	if err != nil {
		ll.Error().Msgf("failed to set bot's ready state: %s", err)
	}
}
