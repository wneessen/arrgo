package bot

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// List of icons/emojis
const (
	IconGold        = "\U0001F7E1"
	IconDoubloon    = "🔵"
	IconAncientCoin = "💰"
	IconIncrease    = "📈 "
	IconDecrease    = "📉 "
	/*
	   "Kraken":      "🐙",
	   "Megalodon":   "🦈",
	   "Chest":       "🗝️",
	   "Ship":        "⛵",
	   "Vomit":       "🤮",
	*/
)

// SoTUserBalance represents the JSON structure of the Sea of Thieves user balance API response
type SoTUserBalance struct {
	GamerTag     string `json:"gamertag"`
	Title        string `json:"title"`
	Doubloons    int    `json:"doubloons"`
	Gold         int    `json:"gold"`
	AncientCoins int    `json:"ancientCoins"`
}

// SlashCmdSoTBalance handles the /balance slash command
func (b *Bot) SlashCmdSoTBalance(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	r := &Requester{i.Member, b.Model.User}
	ub, err := b.SoTGetUserBalance(r)
	if err != nil {
		return err
	}

	p := message.NewPrinter(language.German)
	var ef []*discordgo.MessageEmbedField
	ef = append(ef, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Gold", IconGold),
		Value:  fmt.Sprintf("**%s** Gold", p.Sprintf("%d", ub.Gold)),
		Inline: true,
	})
	ef = append(ef, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Doubloons", IconDoubloon),
		Value:  fmt.Sprintf("**%s** Doubloons", p.Sprintf("%d", ub.Doubloons)),
		Inline: true,
	})
	ef = append(ef, &discordgo.MessageEmbedField{
		Name:   fmt.Sprintf("%s Ancient Coins", IconAncientCoin),
		Value:  fmt.Sprintf("**%s** Ancient Coins", p.Sprintf("%d", ub.AncientCoins)),
		Inline: true,
	})

	e := []*discordgo.MessageEmbed{
		{
			Title:       "Your current balance in Sea of Thieves:",
			Description: fmt.Sprintf("**Title:** %s", ub.Title),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "https://github.com/wneessen/arrgo/raw/main/assets/season/gold-s.png",
			},
			Type:   discordgo.EmbedTypeRich,
			Fields: ef,
		},
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: e}); err != nil {
		return err
	}
	return nil
}

// SoTGetUserBalance returns the parsed API response from the Sea of Thieves gold/coins balance API
func (b *Bot) SoTGetUserBalance(rq *Requester) (SoTUserBalance, error) {
	var ub SoTUserBalance
	c, err := rq.GetSoTRATCookie()
	if err != nil {
		return ub, err
	}
	r, err := b.HTTPClient.HttpReq(ApiURLSoTUserBalance, ReqMethodGet, nil)
	if err != nil {
		return ub, err
	}
	r.SetSOTRequest(c)
	rd, _, err := b.HTTPClient.Fetch(r)
	if err != nil {
		return ub, err
	}
	if err := json.Unmarshal(rd, &ub); err != nil {
		return ub, err
	}
	return ub, nil
}
