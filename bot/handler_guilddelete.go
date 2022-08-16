package bot

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/wneessen/arrgo/model"
)

// GuildDelete receives GUILD_DELETE updates from each server the bot is connected to
func (b *Bot) GuildDelete(_ *discordgo.Session, ev *discordgo.GuildDelete) {
	ll := b.Log.With().Str("context", "bot.GuildDelete").Str("guild_id", ev.Guild.ID).Logger()
	ll.Info().Msgf("received a GUILD_DELETE event... removing from database")
	g, err := b.Model.Guild.GetByGuildID(ev.Guild.ID)
	if err != nil {
		if !errors.Is(err, model.ErrGuildNotExistent) {
			ll.Error().Msgf("failed to fetch guild from DB: %s", err)
			return
		}
		ll.Warn().Msgf("guild not found in database... skipping removal")
		return
	}
	if err := b.Model.Guild.Delete(g); err != nil {
		ll.Error().Msgf("failed to remove guild from database: %s", err)
	}
	ll.Info().Msg("guild successfully removed from database")
}
