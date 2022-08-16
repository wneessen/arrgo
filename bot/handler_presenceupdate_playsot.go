package bot

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/wneessen/arrgo/model"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"time"
)

// UserPlaySoT receives PRESENCE_UPDATE from each server and handles if the user starts playing SoT
func (b *Bot) UserPlaySoT(_ *discordgo.Session, ev *discordgo.PresenceUpdate) {
	ll := b.Log.With().Str("context", "bot.UserPlaySoT").Str("user_id", ev.User.ID).Logger()

	u, err := b.Model.User.GetByUserID(ev.User.ID)
	if err != nil {
		if !errors.Is(err, model.ErrUserNotExistent) {
			ll.Error().Msgf("failed to monitor gaming since user couldn't be retieved from DB: %s", err)
		}
		return
	}
	r := Requester{nil, b.Model.User, u}

	// We want to only monitor Sea of Thieves gaming activities
	ig := false
	for _, a := range ev.Activities {
		if a.Type == discordgo.ActivityTypeGame && a.Name == "Sea of Thieves" {
			ig = true
			break
		}
	}

	// User started playing Sea of Thieves
	if ig {
		wp, err := b.Model.User.GetPrefBool(u, model.UserPrefPlaysSoT)
		if err != nil && !errors.Is(err, model.ErrUserPrefNotExistent) {
			ll.Warn().Msgf(ErrFailedRetrieveUserStatsDB, err)
			return
		}

		// User is already marked as playing
		if wp {
			return
		}

		if _, err := r.GetSoTRATCookie(); err != nil {
			ll.Warn().Msgf("unable to retrieve user's RAT cookie: %s", err)
			return
		}
		ll.Debug().Msg("user started playing Sea of Thieves")
		if err := b.Model.User.SetPref(u, model.UserPrefPlaysSoT, true); err != nil {
			ll.Warn().Msgf("failed to set user's status in database: %s", err)
			return
		}
		if err := b.Model.User.SetPref(u, model.UserPrefPlaysSoTStartTime, time.Now().Unix()); err != nil {
			ll.Warn().Msgf("failed to set user's start time in database: %s", err)
			return
		}
		if err := b.StoreSoTUserStats(u); err != nil {
			ll.Warn().Msgf("failed to store current user stats in DB: %s", err)
			return
		}
	}

	// User likely stopped playing Sea of Thieves
	if !ig {
		wp, err := b.Model.User.GetPrefBool(u, model.UserPrefPlaysSoT)
		if err != nil && !errors.Is(err, model.ErrUserPrefNotExistent) {
			ll.Warn().Msgf(ErrFailedRetrieveUserStatsDB, err)
			return
		}

		// User wasn't playing SoT
		if !wp {
			return
		}

		ll.Debug().Msg("user stopped playing Sea of Thieves")
		if err := b.Model.User.SetPref(u, model.UserPrefPlaysSoT, false); err != nil {
			ll.Warn().Msgf("failed to set user's status in database: %s", err)
			return
		}
		st, err := b.Model.User.GetPrefInt64(u, model.UserPrefPlaysSoTStartTime)
		if err != nil {
			ll.Warn().Msgf("failed to retrieve start time from DB: %s", err)
			return
		}
		et := time.Now().Unix()

		go func(s, e int64, rq *Requester, pu *discordgo.PresenceUpdate) {
			time.Sleep(time.Minute * 1)
			wp, err := b.Model.User.GetPrefBool(rq.User, model.UserPrefPlaysSoT)
			if err != nil && !errors.Is(err, model.ErrUserPrefNotExistent) {
				ll.Warn().Msgf(ErrFailedRetrieveUserStatsDB, err)
				return
			}
			if wp {
				ll.Debug().Msgf("user apparently resumed playing...")
				return
			}
			sto := time.Unix(s, 0)
			eto := time.Unix(e, 0)
			pt := e - s
			if pt < 180 {
				ll.Debug().Msgf("user played less then 3 minutes (%d seconds). There is no chance of "+
					"any changes to the stats", pt)
				return
			}
			if _, err := r.GetSoTRATCookie(); err != nil {
				ll.Warn().Msgf("unable to retrieve user's RAT cookie: %s", err)
				return
			}
			if err := b.StoreSoTUserStats(u); err != nil {
				ll.Warn().Msgf("failed to store current user stats in DB: %s", err)
				return
			}
			uss, err := b.Model.UserStats.GetByUserIDAtTime(r.User.ID, sto)
			if err != nil {
				ll.Warn().Msgf("failed to read start time user stats from DB: %s", err)
				return
			}
			use, err := b.Model.UserStats.GetByUserIDAtTime(r.User.ID, eto)
			if err != nil {
				ll.Warn().Msgf("failed to read end time user stats from DB: %s", err)
				return
			}

			p := message.NewPrinter(language.German)
			var ef []*discordgo.MessageEmbedField
			if uss.Gold != use.Gold {
				v := use.Gold - uss.Gold
				ef = append(ef, &discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("%s Gold", IconGold),
					Value:  fmt.Sprintf("%s **%s** Gold", changeIcon(v), p.Sprintf("%d", v)),
					Inline: true,
				})
			}
			if uss.Doubloons != use.Doubloons {
				v := use.Doubloons - uss.Doubloons
				ef = append(ef, &discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("%s Doubloons", IconDoubloon),
					Value:  fmt.Sprintf("%s **%s** Doubloons", changeIcon(v), p.Sprintf("%d", v)),
					Inline: true,
				})
			}
			if uss.AncientCoins != use.AncientCoins {
				v := use.AncientCoins - uss.AncientCoins
				ef = append(ef, &discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("%s Ancient Coins", IconAncientCoin),
					Value:  fmt.Sprintf("%s **%s** Ancient Coints", changeIcon(v), p.Sprintf("%d", v)),
					Inline: true,
				})
			}
			if uss.KrakenDefeated != use.KrakenDefeated {
				v := use.KrakenDefeated - uss.KrakenDefeated
				ef = append(ef, &discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("%s Kraken", IconKraken),
					Value:  fmt.Sprintf("**%s** defeated", p.Sprintf("%d", v)),
					Inline: true,
				})
			}
			if uss.MegalodonEnounter != use.MegalodonEnounter {
				v := use.MegalodonEnounter - uss.MegalodonEnounter
				ef = append(ef, &discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("%s Megalodon", IconMegalodon),
					Value:  fmt.Sprintf("**%s** encounter(s)", p.Sprintf("%d", v)),
					Inline: true,
				})
			}
			if uss.ChestsHandedIn != use.ChestsHandedIn {
				v := use.ChestsHandedIn - uss.ChestsHandedIn
				ef = append(ef, &discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("%s Chests", IconChest),
					Value:  fmt.Sprintf("**%s** handed in", p.Sprintf("%d", v)),
					Inline: true,
				})
			}
			if uss.ShipsSunk != use.ShipsSunk {
				v := use.ShipsSunk - uss.ShipsSunk
				ef = append(ef, &discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("%s Other Ships", IconShip),
					Value:  fmt.Sprintf("**%s** sunk", p.Sprintf("%d", v)),
					Inline: true,
				})
			}
			if uss.VomittedTimes != use.VomittedTimes {
				v := use.VomittedTimes - uss.VomittedTimes
				ef = append(ef, &discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("%s Vomitted", IconVomit),
					Value:  fmt.Sprintf("**%s** times", p.Sprintf("%d", v)),
					Inline: true,
				})
			}
			if uss.DistanceSailed != use.DistanceSailed {
				v := use.DistanceSailed - uss.DistanceSailed
				ef = append(ef, &discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("%s Distance", IconDistance),
					Value:  fmt.Sprintf("**%s** nmi sailed", p.Sprintf("%d", v)),
					Inline: true,
				})
			}
			for len(ef)%3 != 0 {
				ef = append(ef, &discordgo.MessageEmbedField{
					Value:  "\U0000FEFF",
					Name:   "\U0000FEFF",
					Inline: true,
				})
			}

			du, err := b.Session.User(u.UserID)
			if err != nil {
				ll.Warn().Msgf("failed to retrieve user information from Discord: %s", err)
				return
			}
			eb := []*discordgo.MessageEmbed{
				{
					Title:  fmt.Sprintf("Sea of Thieves voyage summary for @%s", du.Username),
					Type:   discordgo.EmbedTypeRich,
					Fields: ef,
				},
			}

			g, err := b.Model.Guild.GetByGuildID(pu.GuildID)
			if err != nil {
				ll.Error().Msgf("failed to retrieve guild information from DB: %s", err)
				return
			}
			ag, err := b.Model.Guild.GetPrefBool(g, model.GuildPrefAnnounceSoTSummary)
			if err != nil && !errors.Is(err, model.ErrGuildPrefNotExistent) {
				ll.Error().Msgf("failed to fetch guild preference from DB: %s", err)
				return
			}
			if ag {
				if _, err := b.Session.ChannelMessageSendEmbed(b.Model.Guild.AnnouceChannel(g), eb[0]); err != nil {
					ll.Error().Msgf("failed to send voyage summary message: %s", err)
				}
			}
		}(st, et, &r, ev)
	}
}
