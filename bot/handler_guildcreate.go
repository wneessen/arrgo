package bot

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/wneessen/arrgo/config"
	"github.com/wneessen/arrgo/crypto"
	"github.com/wneessen/arrgo/model"
)

// GuildCreate receives GUILD_CREATE updates from each server the bot is connected to
func (b *Bot) GuildCreate(s *discordgo.Session, ev *discordgo.GuildCreate) {
	ll := b.Log.With().Str("context", "bot.GuildCreate").Str("guild_id", ev.Guild.ID).Logger()

	// Check if guild is already present in database
	var g *model.Guild
	var err error
	_, err = b.Model.Guild.GetByGuildID(ev.Guild.ID)
	if err != nil {
		if !errors.Is(err, model.ErrGuildNotExistent) {
			ll.Error().Msgf("failed to fetch guild from DB: %s", err)
			return
		}

		ll.Debug().Msg("guild not found in database... trying to add it")
		gs, err := crypto.RandomBytes(config.CryptoKeyLen)
		if err != nil {
			ll.Error().Msgf("failed to generate guild encryption secret: %s", err)
			return
		}
		ek, err := crypto.EncryptAuth(gs, []byte(b.Config.Data.EncryptionKey), []byte(ev.Guild.ID))
		if err != nil {
			ll.Error().Msgf("failed to encrypt guild encryption secret with global encryption key: %s", err)
			return
		}
		g = &model.Guild{
			GuildID:         ev.Guild.ID,
			GuildName:       ev.Guild.Name,
			OwnerID:         ev.Guild.OwnerID,
			JoinedAt:        ev.Guild.JoinedAt,
			SystemChannelID: ev.Guild.SystemChannelID,
			EncryptionKey:   ek,
		}
		if err := b.Model.Guild.Insert(g); err != nil {
			ll.Error().Msgf("failed to insert guild into database: %s", err)
		}

		// By default we don't want FH spam
		if err := b.Model.Guild.SetPref(g, model.GuildPrefScheduledFlameheart, false); err != nil {
			ll.Error().Msgf("failed to set guild preference FH_SPAM in database: %s", err)
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
				URL: fmt.Sprintf(`%s/piratelord_small.png`, AssetsBaseURL),
			},
			Title:  "Avast ye!",
			Fields: ef,
		}
		if _, err := s.ChannelMessageSendEmbed(ev.Guild.SystemChannelID, e); err != nil {
			ll.Error().Msgf("failed to send introcution message: %s", err)
		}
	}
}
