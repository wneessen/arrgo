package bot

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// SoTAllegianceJSON is the nested struct from the Sea of Thieves event hub response
type SoTAllegianceJSON struct {
	Stats []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"stats"`
}

// SoTAllegiance is the struct that represents the parsed data from the API endpoint
type SoTAllegiance struct {
	Allegiance string
	ShipsSunk  int64
	MaxStreak  int64
	TotalGold  int64
}

// SlashCmdSoTAllegiance handles the /allegiance slash command
func (b *Bot) SlashCmdSoTAllegiance(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	eo := i.ApplicationCommandData().Options
	if len(eo) <= 0 {
		return fmt.Errorf("no option given")
	}
	rc, ok := eo[0].Value.(string)
	if !ok {
		return fmt.Errorf("provided option value is not a string")
	}

	re, err := regexp.Compile(`^(?i:guardians|servants)$`)
	if err != nil {
		return err
	}
	ala := re.FindStringSubmatch(rc)
	if len(ala) != 1 {
		return fmt.Errorf("failed to parse value string")
	}
	al := ala[0]

	r, err := b.NewRequester(i.Interaction)
	if err != nil {
		return err
	}

	a, err := b.SoTGetAllegiance(r, al)
	if err != nil {
		return err
	}

	p := message.NewPrinter(language.German)
	var ef []*discordgo.MessageEmbedField
	ef = append(ef, &discordgo.MessageEmbedField{
		Name:   "Ships Sunk",
		Value:  fmt.Sprintf("%s **%d** Total", IconShip, a.ShipsSunk),
		Inline: true,
	})
	ef = append(ef, &discordgo.MessageEmbedField{
		Name:   "Highest Streak",
		Value:  fmt.Sprintf("%s **%d** Ships", IconGauge, a.MaxStreak),
		Inline: true,
	})
	ef = append(ef, &discordgo.MessageEmbedField{
		Name:   "Highest Hourglass Value",
		Value:  fmt.Sprintf("%s **%s** Gold", IconGold, p.Sprintf("%d", a.TotalGold)),
		Inline: true,
	})

	e := []*discordgo.MessageEmbed{
		{
			Title: fmt.Sprintf("Your current allegiance values for the **%s**:", a.Allegiance),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: fmt.Sprintf("%s/allegiance/%s.png", AssetsBaseURL, al),
			},
			Type:   discordgo.EmbedTypeRich,
			Fields: ef,
		},
	}
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &e}); err != nil {
		return err
	}

	return nil
}

// SoTGetAllegiance returns the parsed API response from the Sea of Thieves allegiance API
func (b *Bot) SoTGetAllegiance(rq *Requester, at string) (SoTAllegiance, error) {
	var a SoTAllegiance
	var al SoTAllegianceJSON
	hc, err := NewHTTPClient()
	if err != nil {
		return a, fmt.Errorf(ErrFailedHTTPClient, err)
	}
	c, err := rq.GetSoTRATCookie()
	if err != nil {
		return a, err
	}

	urlfmt := "%s/%s"
	var url string
	switch strings.ToLower(at) {
	case "guardians":
		url = fmt.Sprintf(urlfmt, APIURLSoTAllegiance, "piratelord")
	case "servants":
		url = fmt.Sprintf(urlfmt, APIURLSoTAllegiance, "flameheart")
	default:
		return a, fmt.Errorf("unknown allegiance given")
	}

	r, err := hc.HTTPReq(url, ReqMethodGet, nil)
	if err != nil {
		return a, err
	}
	r.SetSOTRequest(c)
	rd, _, err := hc.Fetch(r)
	if err != nil {
		return a, err
	}

	if err := json.Unmarshal(rd, &al); err != nil {
		return a, err
	}

	switch strings.ToLower(at) {
	case "guardians":
		for _, d := range al.Stats {
			v, err := strconv.ParseInt(d.Value, 10, 64)
			if err != nil {
				return a, fmt.Errorf(ErrFailedStringConvert, d.Value)
			}
			switch d.Name {
			case "FactionG_Ships_Sunk":
				a.ShipsSunk = v
			case "PirateLord_MaxStreak":
				a.MaxStreak = v
			case "FactionG_SandsOfFate_TotalGold":
				a.TotalGold = v
			}
			a.Allegiance = "Guardians of Fortune"
		}
	case "servants":
		for _, d := range al.Stats {
			if d.Value == "" {
				d.Value = "0"
			}
			v, err := strconv.ParseInt(d.Value, 10, 64)
			if err != nil {
				return a, fmt.Errorf(ErrFailedStringConvert, d.Value)
			}
			switch d.Name {
			case "FactionB_Ships_Sunk":
				a.ShipsSunk = v
			case "Flameheart_MaxStreak":
				a.MaxStreak = v
			case "FactionB_SandsOfFate_TotalGold":
				a.TotalGold = v
			}
		}
		a.Allegiance = "Servants of the Flame"
	}

	return a, nil
}
