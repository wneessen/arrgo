package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/wneessen/arrgo/crypto"
	"github.com/wneessen/arrgo/model"

	"github.com/bwmarrin/discordgo"
)

// SoTReputation represents first level of the JSON structure of the Sea of Thieves reputation
// within a season API response
type SoTReputation map[string]SoTFactionReputation

// SoTFactionReputation represents second level of the JSON structure of the Sea of Thieves reputation
// within a season API response
type SoTFactionReputation struct {
	Name             string
	Motto            string              `json:"Motto"`
	Rank             string              `json:"Rank"`
	Level            int64               `json:"Level"`
	Experience       int64               `json:"XP"`
	NextCompanyLevel SoTFactionNextLevel `json:"NextCompanyLevel"`
	TitlesTotal      int64               `json:"TitlesTotal"`
	TitlesUnlocked   int64               `json:"TitlesUnlocked"`
	EmblemsTotal     int64               `json:"EmblemsTotal"`
	EmblemsUnlocked  int64               `json:"EmblemsUnlocked"`
	ItemsTotal       int64               `json:"ItemsTotal"`
	ItemsUnlocked    int64               `json:"ItemsUnlocked"`
}

// SoTFactionNextLevel represents XP level information of the JSON structure of the Sea of Thieves reputation
// within a season API response
type SoTFactionNextLevel struct {
	Level      int64 `json:"Level"`
	XPRequired int64 `json:"XpRequiredToAttain"`
}

// SlashCmdSoTReputation handles the /reputation slash command
func (b *Bot) SlashCmdSoTReputation(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	fo := i.ApplicationCommandData().Options
	if len(fo) <= 0 {
		return fmt.Errorf("no option given")
	}
	rc, ok := fo[0].Value.(string)
	if !ok {
		return fmt.Errorf("provided option value is not a string")
	}

	re, err := regexp.Compile(`^(?i:factiong|hunterscall|merchantalliance|bilgerats|talltales|athenasfortune|` +
		`goldhoarders|orderofsouls|reapersbones)$`)
	if err != nil {
		return err
	}
	faa := re.FindStringSubmatch(rc)
	if len(faa) != 1 {
		return fmt.Errorf("failed to parse value string")
	}
	fa := faa[0]
	_ = fa

	r, err := b.NewRequester(i.Interaction)
	if err != nil {
		return err
	}

	/*
		rp, err := b.SoTGetReputation(r)
		if err != nil {
			return err
		}
	*/
	b.Log.Debug().Msgf("UID: %d, Emi: %s", r.User.ID, fa)
	ur, err := b.Model.UserReputation.GetByUserIDAtTime(r.User.ID, fa, time.Now().Add(time.Minute*-40))
	if err != nil {
		return err
	}
	b.Log.Debug().Msgf("FACTIONS: %+v\n", ur)

	/*
		var ef []*discordgo.MessageEmbedField
		ef = append(ef, &discordgo.MessageEmbedField{
			Name:   "Faction/Company",
			Value:  l.Name,
			Inline: false,
		})
		ef = append(ef, &discordgo.MessageEmbedField{
			Name:   "Current Title",
			Value:  l.BandTitle,
			Inline: false,
		})
		ef = append(ef, &discordgo.MessageEmbedField{
			Name:   "Emissary value",
			Value:  fmt.Sprintf("%s **%d**", IconAncientCoin, l.Score),
			Inline: true,
		})
		ef = append(ef, &discordgo.MessageEmbedField{
			Name:   "Ledger position",
			Value:  fmt.Sprintf("%s **%d**", IconGauge, l.Rank),
			Inline: true,
		})
		ef = append(ef, &discordgo.MessageEmbedField{
			Name:   "Next level in",
			Value:  fmt.Sprintf("%s **%d** points", IconIncrease, l.ToNextRank),
			Inline: true,
		})

		e := []*discordgo.MessageEmbed{
			{
				Title: "Your global ledger in Sea of Thieves:",
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: fmt.Sprintf("%s/ledger/%s%d.png", AssetsBaseURL, fa, 4-l.Band),
				},
				Type:   discordgo.EmbedTypeRich,
				Fields: ef,
			},
		}
		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &e}); err != nil {
			return err
		}

	*/
	return nil
}

// SoTGetReputation returns the parsed API response from the Sea of Thieves reputation API
func (b *Bot) SoTGetReputation(rq *Requester) (SoTReputation, error) {
	var re SoTReputation
	hc, err := NewHTTPClient()
	if err != nil {
		return re, fmt.Errorf(ErrFailedHTTPClient, err)
	}
	c, err := rq.GetSoTRATCookie()
	if err != nil {
		return re, err
	}
	r, err := hc.HTTPReq(APIURLSoTReputation, ReqMethodGet, nil)
	if err != nil {
		return re, err
	}
	r.SetSOTRequest(c)
	rd, ho, err := hc.Fetch(r)
	if err != nil {
		return re, err
	}
	if ho.StatusCode == http.StatusUnauthorized {
		return re, ErrSOTUnauth
	}
	if err := json.Unmarshal(rd, &re); err != nil {
		return re, err
	}
	return re, nil
}

// StoreSoTUserReputation will retrieve the latest user reputation from the API and store them in the DB
func (b *Bot) StoreSoTUserReputation(u *model.User) error {
	r, err := NewRequesterFromUser(u, b.Model.User)
	if err != nil {
		b.Log.Warn().Msgf("failed to create new requester: %s", err)
		return err
	}
	ur, err := b.SoTGetReputation(r)
	if err != nil {
		switch {
		case errors.Is(err, ErrSOTUnauth):
			b.Log.Warn().Msgf("failed to fetch user reputation - RAT token is expired")
			return nil
		default:
			return fmt.Errorf("failed to fetch user reputation for user %s: %w", u.UserID, err)
		}
	}
	for k, rep := range ur {
		dur := &model.UserReputation{
			UserID:              u.ID,
			Emissary:            k,
			Motto:               rep.Motto,
			Rank:                rep.Rank,
			Level:               rep.Level,
			Experience:          rep.Experience,
			NextLevel:           rep.NextCompanyLevel.Level,
			ExperienceNextLevel: rep.NextCompanyLevel.XPRequired,
			TitlesTotal:         rep.TitlesTotal,
			TitlesUnlocked:      rep.TitlesUnlocked,
			EmblemsTotal:        rep.EmblemsTotal,
			EmblemsUnlocked:     rep.EmblemsUnlocked,
			ItemsTotal:          rep.ItemsTotal,
			ItemsUnlocked:       rep.ItemsUnlocked,
		}
		if err := b.Model.UserReputation.Insert(dur); err != nil {
			return fmt.Errorf("failed to store user reputation for user %q in DB: %w", u.UserID, err)
		}
	}
	return nil
}

// ScheduledEventUpdateUserReputation performs scheuled updates of the SoT user reputation for each user
func (b *Bot) ScheduledEventUpdateUserReputation() error {
	ll := b.Log.With().Str("context", "bot.ScheduledEventUpdateUserReputation").Logger()
	ul, err := b.Model.User.GetUsers()
	if err != nil {
		return fmt.Errorf("failed to retrieve user list from DB: %w", err)
	}
	for _, u := range ul {
		if err := b.StoreSoTUserReputation(u); err != nil {
			ll.Error().Msgf("failed to store user reputation in DB: %s", err)
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
