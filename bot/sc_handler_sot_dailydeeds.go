package bot

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/wneessen/arrgo/model"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"regexp"
	"time"
)

// SoTEventHubJSON is the nested struct from the Sea of Thieves event hub response
type SoTEventHubJSON struct {
	Data struct {
		Components []struct {
			Data struct {
				BountyList []SoTDeed `json:"BountyList"`
			} `json:"data"`
		} `json:"components"`
	} `json:"data"`
}

// SoTDeed is a deed as returned by the events-hub in Sea of Thieves
type SoTDeed struct {
	Type         string          `json:"#Type"`
	Title        string          `json:"Title"`
	BodyText     string          `json:"BodyText"`
	StartDateApi *APITimeRFC3339 `json:"StartDate,omitempty"`
	EndDateApi   *APITimeRFC3339 `json:"EndDate,omitempty"`
	Image        struct {
		Desktop string `json:"desktop"`
	} `json:"Image"`
	RewardDetails struct {
		Gold      int    `json:"Gold"`
		Doubloons int    `json:"Doubloons"`
		XPGain    string `json:"XPGain"`
	} `json:"RewardDetails"`
}

// SlashCmdSoTDailyDeeds handles the /dailydeed slash command
func (b *Bot) SlashCmdSoTDailyDeeds(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	dl, err := b.Model.Deed.GetByDeedsAtTime(time.Now())
	if err != nil {
		return err
	}
	if len(dl) <= 0 {
		return fmt.Errorf("no deeds found for today in database")
	}

	var e []*discordgo.MessageEmbed
	c := cases.Title(language.English)
	for _, d := range dl {
		var t, rg string
		switch d.DeedType {
		case model.DeedTypeStandard:
			t = "Standard Deed"
		case model.DeedTypeDailyStandard:
			t = "Standard Daily Deed"
		case model.DeedTypeDailySwift:
			t = "Daily Swift Deed"
		}
		switch d.RewardIcon {
		case "s":
			rg = "Small renown"
		case "m":
			rg = "Medium renown"
		}
		de := fmt.Sprintf("%s\n\n**Valid from:** <t:%d>\n**Valid thru:** <t:%d>\n**Reward:** %d %s\n"+
			"**Renown gain:** %s",
			d.Description, d.ValidFrom.Unix(), d.ValidThru.Unix(), d.RewardAmount,
			c.String(string(d.RewardType)), rg)
		if t != "" {
			ce := &discordgo.MessageEmbed{
				Title:       t,
				Description: de,
				Type:        discordgo.EmbedTypeArticle,
				Image: &discordgo.MessageEmbedImage{
					URL: d.ImageURL,
				},
			}
			e = append(e, ce)
		}
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: e}); err != nil {
		return err
	}
	return nil
}

// ScheduledEventUpdateDailyDeeds performs scheuled updates of the SoT daily deeds
func (b *Bot) ScheduledEventUpdateDailyDeeds() error {
	ll := b.Log.With().Str("context", "bot.ScheduledEventUpdateDailyDeeds").Logger()
	dl, err := b.SoTGetDailyDeeds()
	if err != nil {
		return fmt.Errorf("failed to fetch deeds from event hub: %w", err)
	}

	for _, d := range dl {
		dbd := &model.Deed{
			Description: d.BodyText,
			ImageURL:    d.Image.Desktop,
			RewardIcon:  d.RewardDetails.XPGain,
		}
		switch d.Title {
		case "Standard Deed":
			dbd.DeedType = model.DeedTypeStandard
		case "Standard Daily Deed":
			dbd.DeedType = model.DeedTypeDailyStandard
		case "Swift Daily Deed":
			dbd.DeedType = model.DeedTypeDailySwift
		default:
			dbd.DeedType = model.DeedTypeUnknown
		}
		if d.RewardDetails.Gold > 0 {
			dbd.RewardType = model.RewardGold
			dbd.RewardAmount = d.RewardDetails.Gold
		}
		if d.RewardDetails.Doubloons > 0 {
			dbd.RewardType = model.RewardDoubloons
			dbd.RewardAmount = d.RewardDetails.Doubloons
		}
		if d.StartDateApi != nil {
			dbd.ValidFrom = time.Time(*d.StartDateApi)
		}
		if d.EndDateApi != nil {
			dbd.ValidThru = time.Time(*d.EndDateApi)
		}
		if err := b.Model.Deed.Insert(dbd); err != nil && !errors.Is(err, model.ErrDeedDuplicate) {
			ll.Error().Msgf("failed to insert deed into database: %s", err)
		}
	}
	return nil
}

// SoTGetDailyDeeds returns the parsed API response from the Sea of Thieves event-hub API
func (b *Bot) SoTGetDailyDeeds() ([]SoTDeed, error) {
	var dl []SoTDeed
	hc, err := NewHTTPClient()
	if err != nil {
		return dl, fmt.Errorf(ErrFailedHTTPClient, err)
	}

	// We need a valid RAT token first
	ul, err := b.Model.User.GetUsers()
	if err != nil {
		return dl, fmt.Errorf("failed to retrieve user list from DB: %w", err)
	}
	var rc string
	for _, u := range ul {
		rq := &Requester{nil, b.Model.User, u}
		uc, err := rq.GetSoTRATCookie()
		if err != nil {
			b.Log.Debug().Msgf("failed to fetch users RAT cookie: %s", err)
			continue
		}
		if uc != "" {
			rc = uc
			break
		}
	}
	r, err := hc.HttpReq(ApiURLSoTEventHub, ReqMethodGet, nil)
	if err != nil {
		return dl, err
	}
	r.SetSOTRequest(rc)
	rd, _, err := hc.Fetch(r)
	if err != nil {
		return dl, err
	}
	re, err := regexp.Compile(`<script>var APP_PROPS\s*=\s*({.*});</script>`)
	if err != nil {
		return dl, err
	}
	rj := re.FindStringSubmatch(string(rd))
	if len(rj) != 2 {
		return dl, fmt.Errorf("failed to parse API response from SoT events hub")
	}

	var ehj SoTEventHubJSON
	if err := json.Unmarshal([]byte(rj[1]), &ehj); err != nil {
		return dl, err
	}
	dl = ehj.Data.Components[1].Data.BountyList

	return dl, nil
}
