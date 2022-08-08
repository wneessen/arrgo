package bot

import (
	"github.com/bwmarrin/discordgo"
)

// UserPlaySoT receives PRESENCE_UPDATE from each server and handles if the user starts playing SoT
func (b *Bot) UserPlaySoT(_ *discordgo.Session, ev *discordgo.PresenceUpdate) {
	ll := b.Log.With().Str("context", "bot.UserPlaySoT").Str("user_id", ev.User.ID).Logger()
	ll.Info().Msgf("received a PRESENCE_UPDATE event...")
	ll.Debug().Msgf("%+v", ev)
}
