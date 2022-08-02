package bot

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/wneessen/arrgo/model"
)

// GuildCreate receives GUILD_CREATE updates from each server the bot is connected to
func (b *Bot) GuildCreate(_ *discordgo.Session, ev *discordgo.GuildCreate) {
	ll := b.Log.With().Str("context", "bot.GuildCreate").Str("guild_id", ev.Guild.ID).Logger()

	// Check if guild is already present in database
	g, err := b.Model.Guild.GetByGuildID(ev.Guild.ID)
	if err != nil {
		if !errors.Is(err, model.ErrGuildNotExistant) {
			ll.Error().Msgf("failed to fetch guild from DB: %s", err)
			return
		}

		ll.Debug().Msg("guild not found in database... trying to add it")
		g = &model.Guild{
			GuildID:   ev.Guild.ID,
			GuildName: ev.Guild.Name,
			OwnerID:   ev.Guild.OwnerID,
		}
		if err := b.Model.Guild.Insert(g); err != nil {
			ll.Error().Msgf("failed to insert guild into database: %s", err)
		}
	}
	ll.Debug().Msgf("Guild: %+v", g)
}
