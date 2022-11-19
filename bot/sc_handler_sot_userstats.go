package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/wneessen/arrgo/crypto"
	"github.com/wneessen/arrgo/model"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// SoTUserOverview represents the JSON structure of the Sea of Thieves user overview API response
type SoTUserOverview struct {
	Stats SoTUserStats `json:"stats"`
}

// SoTUserStats represents a subpart of the JSON structure of the Sea of Thieves user overview API response
type SoTUserStats struct {
	KrakenDefeated      APIIntString `json:"Combat_Kraken_Defeated"`
	MegalodonEncounters APIIntString `json:"Player_TinyShark_Spawned"`
	ChestsHandedIn      APIIntString `json:"Chests_HandedIn_Total"`
	ShipsSunk           APIIntString `json:"Combat_Ships_Sunk"`
	VomitedTotal        APIIntString `json:"Vomited_Total"`
	MetresSailed        APIIntString `json:"Voyages_MetresSailed_Total"`
}

// SoTUserBalance represents the JSON structure of the Sea of Thieves user balance API response
type SoTUserBalance struct {
	GamerTag     string `json:"gamertag"`
	Title        string `json:"title"`
	Doubloons    int64  `json:"doubloons"`
	Gold         int64  `json:"gold"`
	AncientCoins int64  `json:"ancientCoins"`
}

// SlashCmdSoTOverview handles the /balance slash command
func (b *Bot) SlashCmdSoTOverview(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	u, err := b.Model.User.GetByUserID(i.Member.User.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve user from DB: %w", err)
	}
	if err := b.StoreSoTUserStats(u); err != nil {
		return fmt.Errorf("failed to update user stats in DB: %w", err)
	}
	us, err := b.Model.UserStats.GetByUserID(u.ID)
	if err != nil {
		return err
	}

	p := message.NewPrinter(language.German)
	var ef []*discordgo.MessageEmbedField
	ef = append(ef, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Kraken", IconKraken),
		Value:  fmt.Sprintf("**%s** defeated", p.Sprintf("%d", us.KrakenDefeated)),
		Inline: true,
	})
	ef = append(ef, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Megalodon", IconMegalodon),
		Value:  fmt.Sprintf("**%s** encounter(s)", p.Sprintf("%d", us.MegalodonEnounter)),
		Inline: true,
	})
	ef = append(ef, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Chests", IconChest),
		Value:  fmt.Sprintf("**%s** handed in", p.Sprintf("%d", us.ChestsHandedIn)),
		Inline: true,
	})
	ef = append(ef, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Other Ships", IconShip),
		Value:  fmt.Sprintf("**%s** sunk", p.Sprintf("%d", us.ShipsSunk)),
		Inline: true,
	})
	ef = append(ef, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Vomitted", IconVomit),
		Value:  fmt.Sprintf("**%s** times", p.Sprintf("%d", us.VomittedTimes)),
		Inline: true,
	})
	if us.DistanceSailed > 0 {
		ef = append(ef, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Distance", IconDistance),
			Value:  fmt.Sprintf("**%s** nmi sailed", p.Sprintf("%d", us.DistanceSailed/1852)),
			Inline: true,
		})
	} else {
		ef = append(ef, &discordgo.MessageEmbedField{
			Value:  "\U0000FEFF",
			Name:   "\U0000FEFF",
			Inline: true,
		})
	}

	e := []*discordgo.MessageEmbed{
		{
			Title:  "Your current user statistics overview in Sea of Thieves:",
			Type:   discordgo.EmbedTypeRich,
			Fields: ef,
		},
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &e}); err != nil {
		return err
	}
	return nil
}

// SlashCmdSoTBalance handles the /balance slash command
func (b *Bot) SlashCmdSoTBalance(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	u, err := b.Model.User.GetByUserID(i.Member.User.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve user from DB: %w", err)
	}
	if err := b.StoreSoTUserStats(u); err != nil {
		return fmt.Errorf("failed to update user stats in DB: %w", err)
	}
	ub, err := b.Model.UserStats.GetByUserID(u.ID)
	if err != nil {
		return err
	}

	p := message.NewPrinter(language.German)
	var ef []*discordgo.MessageEmbedField
	ef = append(ef, &discordgo.MessageEmbedField{
		Name: fmt.Sprintf("%s Gold", IconGold),
		Value: fmt.Sprintf("%s **%s** Gold", changeIcon(ub.Gold),
			p.Sprintf("%d", ub.Gold)),
		Inline: true,
	})
	ef = append(ef, &discordgo.MessageEmbedField{
		Name: fmt.Sprintf("%s Doubloons", IconDoubloon),
		Value: fmt.Sprintf("%s **%s** Doubloons", changeIcon(ub.Doubloons),
			p.Sprintf("%d", ub.Doubloons)),
		Inline: true,
	})
	ef = append(ef, &discordgo.MessageEmbedField{
		Name: fmt.Sprintf("%s Ancient Coins", IconAncientCoin),
		Value: fmt.Sprintf("%s **%s** Ancient Coins", changeIcon(ub.AncientCoins),
			p.Sprintf("%d", ub.AncientCoins)),
		Inline: true,
	})

	e := []*discordgo.MessageEmbed{
		{
			Title: "Your current balance in Sea of Thieves:",
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: fmt.Sprintf("%s/season/gold-s.png", AssetsBaseURL),
			},
			Type:   discordgo.EmbedTypeRich,
			Fields: ef,
		},
	}
	if ub.Title != "" {
		e[0].Description = fmt.Sprintf("**Current Title:** %s", ub.Title)
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &e}); err != nil {
		return err
	}
	return nil
}

// SoTGetUserBalance returns the parsed API response from the Sea of Thieves gold/coins balance API
func (b *Bot) SoTGetUserBalance(rq *Requester) (SoTUserBalance, error) {
	var ub SoTUserBalance
	hc, err := NewHTTPClient()
	if err != nil {
		return ub, fmt.Errorf(ErrFailedHTTPClient, err)
	}
	c, err := rq.GetSoTRATCookie()
	if err != nil {
		return ub, err
	}
	r, err := hc.HTTPReq(APIURLSoTUserBalance, ReqMethodGet, nil)
	if err != nil {
		return ub, err
	}
	r.SetSOTRequest(c)
	rd, ho, err := hc.Fetch(r)
	if err != nil {
		return ub, err
	}
	if ho.StatusCode == http.StatusUnauthorized {
		return ub, ErrSOTUnauth
	}
	if err := json.Unmarshal(rd, &ub); err != nil {
		return ub, err
	}
	return ub, nil
}

// SoTGetUserOverview returns the parsed API response from the Sea of Thieves gold/coins balance API
func (b *Bot) SoTGetUserOverview(rq *Requester) (SoTUserStats, error) {
	var us SoTUserOverview
	hc, err := NewHTTPClient()
	if err != nil {
		return SoTUserStats{}, fmt.Errorf(ErrFailedHTTPClient, err)
	}
	c, err := rq.GetSoTRATCookie()
	if err != nil {
		return SoTUserStats{}, err
	}
	r, err := hc.HTTPReq(APIURLSoTUserOverview, ReqMethodGet, nil)
	if err != nil {
		return SoTUserStats{}, err
	}
	r.SetSOTRequest(c)
	rd, _, err := hc.Fetch(r)
	if err != nil {
		return SoTUserStats{}, err
	}
	if err := json.Unmarshal(rd, &us); err != nil {
		return us.Stats, err
	}
	return us.Stats, nil
}

// ScheduledEventUpdateUserStats performs scheuled updates of the SoT user stats for each user
func (b *Bot) ScheduledEventUpdateUserStats() error {
	ll := b.Log.With().Str("context", "bot.ScheduledEventUpdateUserStats").Logger()
	ul, err := b.Model.User.GetUsers()
	if err != nil {
		return fmt.Errorf("failed to retrieve user list from DB: %w", err)
	}
	for _, u := range ul {
		if err := b.StoreSoTUserStats(u); err != nil {
			ll.Error().Msgf("failed to store user stats in DB: %s", err)
			continue
		}
		rd, err := crypto.RandDuration(10, "s")
		if err != nil {
			rd = time.Second * 10
		}
		time.Sleep(rd)
	}
	return nil
}

// StoreSoTUserStats will retrieve the latest user stats from the API and store them in the DB
func (b *Bot) StoreSoTUserStats(u *model.User) error {
	r, err := NewRequesterFromUser(u, b.Model.User)
	if err != nil {
		b.Log.Warn().Msgf("failed to create new requester: %s", err)
		return err
	}
	ub, err := b.SoTGetUserBalance(r)
	if err != nil {
		switch {
		case errors.Is(err, ErrSOTUnauth):
			b.Log.Warn().Msgf("failed to fetch user balance - RAT token is expired")
			return nil
		default:
			return fmt.Errorf("failed to fetch user balance for user %s: %w", u.UserID, err)
		}
	}
	us, err := b.SoTGetUserOverview(r)
	if err != nil {
		return fmt.Errorf("failed to fetch user stats for user %q: %w", u.UserID, err)
	}
	dus := &model.UserStat{
		UserID:            u.ID,
		Title:             ub.Title,
		Gold:              ub.Gold,
		Doubloons:         ub.Doubloons,
		AncientCoins:      ub.AncientCoins,
		KrakenDefeated:    int64(us.KrakenDefeated),
		MegalodonEnounter: int64(us.MegalodonEncounters),
		ChestsHandedIn:    int64(us.ChestsHandedIn),
		ShipsSunk:         int64(us.ShipsSunk),
		VomittedTimes:     int64(us.VomitedTotal),
		DistanceSailed:    int64(us.MetresSailed),
	}
	if err := b.Model.UserStats.Insert(dus); err != nil {
		return fmt.Errorf("failed to store user stats for user %q in DB: %w", u.UserID, err)
	}
	return nil
}
