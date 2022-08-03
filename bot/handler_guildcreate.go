package bot

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/wneessen/arrgo/model"
)

// GuildCreate receives GUILD_CREATE updates from each server the bot is connected to
func (b *Bot) GuildCreate(s *discordgo.Session, ev *discordgo.GuildCreate) {
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
			GuildID:         ev.Guild.ID,
			GuildName:       ev.Guild.Name,
			OwnerID:         ev.Guild.OwnerID,
			JoinedAt:        ev.Guild.JoinedAt,
			SystemChannelID: ev.Guild.SystemChannelID,
		}
		if err := b.Model.Guild.Insert(g); err != nil {
			ll.Error().Msgf("failed to insert guild into database: %s", err)
		}
	}

	// Send introduction to system channel
	ef := []*discordgo.MessageEmbedField{
		{
			Name: "Ahoy, Mateys!",
			Value: "I am ArrGo the Discord Pirate Lord! I just joined this nice vessel to have " +
				"an an eye on you scallywags!",
			Inline: true,
		},
	}
	e := &discordgo.MessageEmbed{
		Type: discordgo.EmbedTypeRich,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: `https://github.com/wneessen/arrgo/raw/main/assets/piratelord_small.png`,
		},
		Title:  "Avast ye!",
		Fields: ef,
	}
	if _, err := s.ChannelMessageSendEmbed(ev.Guild.SystemChannelID, e); err != nil {
		ll.Error().Msgf("failed to send introcution message: %s", err)
	}
}
