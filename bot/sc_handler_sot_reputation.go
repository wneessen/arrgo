package bot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

// SoTReputation represents first level of the JSON structure of the Sea of Thieves reputation
// within a season API response
type SoTReputation struct {
	AthenasFortune     SoTFactionReputation `json:"AthenasFortune"`
	HuntersCall        SoTFactionReputation `json:"HuntersCall"`
	GoldHoarders       SoTFactionReputation `json:"GoldHoarders"`
	OrderOfSouls       SoTFactionReputation `json:"OrderOfSouls"`
	MerchantAlliance   SoTFactionReputation `json:"MerchantAlliance"`
	CreatorCrew        SoTFactionReputation `json:"CreatorCrew"`
	BilgeRats          SoTFactionReputation `json:"BilgeRats"`
	TallTales          SoTFactionReputation `json:"TallTales"`
	ReapersBones       SoTFactionReputation `json:"ReapersBones"`
	ServantsOfTheFlame SoTFactionReputation `json:"FactionB"`
	GuardiansOfFortune SoTFactionReputation `json:"FactionG"`
}

// SoTFactionReputation represents second level of the JSON structure of the Sea of Thieves reputation
// within a season API response
type SoTFactionReputation struct {
	Name             string
	Motto            string              `json:"Motto"`
	Rank             string              `json:"Rank"`
	Level            int                 `json:"Level"`
	Experience       int64               `json:"XP"`
	NextCompanyLevel SoTFactionNextLevel `json:"NextCompanyLevel"`
	TitlesTotal      int                 `json:"TitlesTotal"`
	TitlesUnlocked   int                 `json:"TitlesUnlocked"`
	EmblemsTotal     int                 `json:"EmblemsTotal"`
	EmblemsUnlocked  int                 `json:"EmblemsUnlocked"`
	ItemsTotal       int                 `json:"ItemsTotal"`
	ItemsUnlocked    int                 `json:"ItemsUnlocked"`
}

// SoTFactionNextLevel represents XP level information of the JSON structure of the Sea of Thieves reputation
// within a season API response
type SoTFactionNextLevel struct {
	Level      int `json:"Level"`
	XPRequired int `json:"XpRequiredToAttain"`
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

	re, err := regexp.Compile(`^(?i:athena|hoarder|merchant|order|reaper|hunter|servants|guardians)$`)
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

	rp, err := b.SoTGetReputation(r)
	if err != nil {
		return err
	}
	b.Log.Debug().Msgf("FACTIONS: %+v\n", rp)

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
