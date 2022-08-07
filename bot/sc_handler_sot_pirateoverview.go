package bot

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
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

// SlashCmdSoTOverview handles the /balance slash command
func (b *Bot) SlashCmdSoTOverview(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	r := &Requester{i.Member, b.Model.User}
	us, err := b.SoTGetUserOverview(r)
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
		Value:  fmt.Sprintf("**%s** encounter(s)", p.Sprintf("%d", us.MegalodonEncounters)),
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
		Value:  fmt.Sprintf("**%s** times", p.Sprintf("%d", us.VomitedTotal)),
		Inline: true,
	})
	if us.MetresSailed > 0 {
		ef = append(ef, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Distance", IconDistance),
			Value:  fmt.Sprintf("**%s** nmi sailed", p.Sprintf("%d", us.MetresSailed/1852)),
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

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: e}); err != nil {
		return err
	}
	return nil
}

// SoTGetUserOverview returns the parsed API response from the Sea of Thieves gold/coins balance API
func (b *Bot) SoTGetUserOverview(rq *Requester) (SoTUserStats, error) {
	var us SoTUserOverview
	c, err := rq.GetSoTRATCookie()
	if err != nil {
		return us.Stats, err
	}
	r, err := b.HTTPClient.HttpReq(ApiURLSoTUserOverview, ReqMethodGet, nil)
	if err != nil {
		return us.Stats, err
	}
	r.SetSOTRequest(c)
	rd, _, err := b.HTTPClient.Fetch(r)
	if err != nil {
		return us.Stats, err
	}
	if err := json.Unmarshal(rd, &us); err != nil {
		return us.Stats, err
	}
	return us.Stats, nil
}
