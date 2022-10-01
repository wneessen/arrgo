package bot

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"regexp"
	"strings"
)

// SoTLedger represents the JSON structure of the Sea of Thieves leder positions within a season API response
type SoTLedger struct {
	Current SoTCurrentLedger `json:"current"`
}

// SoTCurrentLedger represents the JSON structure of the Sea of Thieves current leder within the overall
// ledger response
type SoTCurrentLedger struct {
	Friends SoTFriendsLedger `json:"friends"`
}

// SoTFriendsLedger represents the JSON structure of the Sea of Thieves friends positioning withing the current leder
type SoTFriendsLedger struct {
	User SoTEmissaryLedger `json:"user"`
}

// SoTEmissaryLedger represents the JSON structure of the Sea of Thieves ledger data
type SoTEmissaryLedger struct {
	Name       string
	Band       int `json:"band"`
	BandTitle  string
	Rank       int `json:"rank"`
	Score      int `json:"score"`
	ToNextRank int `json:"toNextRank"`
}

// SlashCmdSoTLedger handles the /ledger slash command
func (b *Bot) SlashCmdSoTLedger(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	eo := i.ApplicationCommandData().Options
	if len(eo) <= 0 {
		return fmt.Errorf("no option given")
	}
	rc, ok := eo[0].Value.(string)
	if !ok {
		return fmt.Errorf("provided option value is not a string")
	}

	re, err := regexp.Compile(`^(?i:athena|hoarder|merchant|order|reaper)$`)
	if err != nil {
		return err
	}
	ema := re.FindStringSubmatch(rc)
	if len(ema) != 1 {
		return fmt.Errorf("failed to parse value string")
	}
	em := ema[0]

	r, err := b.NewRequester(i.Interaction)
	if err != nil {
		return err
	}

	l, err := b.SoTGetLedger(r, em)
	if err != nil {
		return err
	}

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
				URL: fmt.Sprintf("%s/ledger/%s%d.png", AssetsBaseURL, em, 4-l.Band),
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

// SoTGetLedger returns the parsed API response from the Sea of Thieves leaderboard ledger API
func (b *Bot) SoTGetLedger(rq *Requester, em string) (SoTEmissaryLedger, error) {
	var l SoTEmissaryLedger
	var al SoTLedger
	hc, err := NewHTTPClient()
	if err != nil {
		return l, fmt.Errorf(ErrFailedHTTPClient, err)
	}
	c, err := rq.GetSoTRATCookie()
	if err != nil {
		return l, err
	}

	urlfmt := "%s/%s?count=3"
	var url string
	switch strings.ToLower(em) {
	case "athena":
		url = fmt.Sprintf(urlfmt, ApiURLSoTLedger, "AthenasFortune")
	case "hoarder":
		url = fmt.Sprintf(urlfmt, ApiURLSoTLedger, "GoldHoarders")
	case "merchant":
		url = fmt.Sprintf(urlfmt, ApiURLSoTLedger, "MerchantAlliance")
	case "order":
		url = fmt.Sprintf(urlfmt, ApiURLSoTLedger, "OrderOfSouls")
	case "reaper":
		url = fmt.Sprintf(urlfmt, ApiURLSoTLedger, "ReapersBones")
	default:
		return l, fmt.Errorf("unknown emissary given")
	}

	r, err := hc.HttpReq(url, ReqMethodGet, nil)
	if err != nil {
		return l, err
	}
	r.SetSOTRequest(c)
	rd, _, err := hc.Fetch(r)
	if err != nil {
		return l, err
	}

	if err := json.Unmarshal(rd, &al); err != nil {
		return l, err
	}

	l = al.Current.Friends.User
	switch strings.ToLower(em) {
	case "athena":
		tl := []string{"Legend", "Guardian", "Voyager", "Seeker"}
		l.Name = "Athena's Fortune"
		l.BandTitle = tl[l.Band]
	case "hoarder":
		tl := []string{"Captain", "Marauder", "Seafarer", "Castaway"}
		l.Name = "Gold Hoarders"
		l.BandTitle = tl[l.Band]
	case "merchant":
		tl := []string{"Admiral", "Commander", "Cadet", "Sailor"}
		l.Name = "Merchant Alliance"
		l.BandTitle = tl[l.Band]
	case "order":
		tl := []string{"Grandee", "Chief", "Mercenary", "Apprentice"}
		l.Name = "Order of Souls"
		l.BandTitle = tl[l.Band]
	case "reaper":
		tl := []string{"Master", "Keeper", "Servant", "Follower"}
		l.Name = "Reaper's Bones"
		l.BandTitle = tl[l.Band]
	}

	return l, nil
}
