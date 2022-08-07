package bot

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/wneessen/arrgo/model"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"regexp"
	"strings"
	"time"
)

// RTTraderoute represents the JSON structure of the rarethief.com traderoute API response
type RTTraderoute struct {
	Dates     string             `json:"trade_route_dates"`
	Routes    map[string]RTRoute `json:"routes"`
	ValidFrom time.Time
	ValidThru time.Time
}

// RTRoute represents the JSON structure of a specific route within the rarethief.com traderoutes
// API response
type RTRoute struct {
	Outpost     string `json:"outpost"`
	SoughtAfter string `json:"sought_after"`
	Surplus     string `json:"surplus"`
}

// SlashCmdSoTTradeRoutes handles the /balance slash command
func (b *Bot) SlashCmdSoTTradeRoutes(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	tl, err := b.Model.TradeRoute.GetTradeRoutes()
	if err != nil {
		return err
	}

	var ef []*discordgo.MessageEmbedField
	c := cases.Title(language.English)
	for _, tr := range tl {
		ef = append(ef, &discordgo.MessageEmbedField{
			Name: tr.Outpost,
			Value: fmt.Sprintf("%s **%s**\n%s **%s**", IconArrowUp, c.String(tr.Surplus),
				IconArrowDown, c.String(tr.SoughtAfter)),
			Inline: true,
		})

	}

	e := []*discordgo.MessageEmbed{
		{
			Title:       "Trade Routes",
			Description: fmt.Sprintf("valid thru %s", tl[0].ValidThru.Format(time.RFC1123)),
			Type:        discordgo.EmbedTypeRich,
			Footer:      &discordgo.MessageEmbedFooter{Text: "Source: https://maps.seaofthieves.rarethief.com/"},
			Fields:      ef,
		},
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: e}); err != nil {
		return err
	}
	return nil
}

// ScheduledEventUpdateTradeRoutes performs scheuled updates of the TR data from rarethief.com
func (b *Bot) ScheduledEventUpdateTradeRoutes() error {
	ll := b.Log.With().Str("context", "bot.ScheduledEventUpdateTradeRoutes").Logger()
	rl, err := b.Model.TradeRoute.GetTradeRoutes()
	if err != nil {
		return fmt.Errorf("failed to retrieve trade routes list from DB: %w", err)
	}
	if len(rl) > 0 {
		dbv, err := b.Model.TradeRoute.ValidThru()
		if err != nil {
			return fmt.Errorf("failed to retrieve trade routes validity date from DB: %w", err)
		}
		if dbv.Unix() > time.Now().Unix() {
			ll.Debug().Msgf("trade routes in DB are still valid. Skipping update")
			return nil
		}
	}
	tr, err := b.RTGetTradeRoutes()
	if err != nil {
		return fmt.Errorf("failed to fetch traderoute: %w", err)
	}
	for _, r := range tr.Routes {
		dtr, err := b.Model.TradeRoute.GetByOutpost(r.Outpost)
		dtr.Outpost = r.Outpost
		dtr.SoughtAfter = r.SoughtAfter
		dtr.Surplus = r.Surplus
		dtr.ValidThru = tr.ValidThru
		if err != nil {
			if !errors.Is(err, model.ErrTradeRouteNotExistant) {
				ll.Error().Msgf("failed to retrieve trade route for %q from DB: %s", r.Outpost, err)
				continue
			}
			if err := b.Model.TradeRoute.Insert(dtr); err != nil {
				ll.Error().Msgf("failed to insert trade route for %q into DB: %s", r.Outpost, err)
				continue
			}
		}
		if err := b.Model.TradeRoute.Update(dtr); err != nil {
			ll.Error().Msgf("failed to update trade route for %q into DB: %s", r.Outpost, err)
		}
	}
	return nil
}

// RTGetTradeRoutes returns the parsed API response from the rarethief.com traderoutes API
func (b *Bot) RTGetTradeRoutes() (RTTraderoute, error) {
	var tr RTTraderoute
	r, err := b.HTTPClient.HttpReq(ApiURLRTTradeRoutes, ReqMethodGet, nil)
	if err != nil {
		return tr, err
	}
	rd, _, err := b.HTTPClient.Fetch(r)
	if err != nil {
		return tr, err
	}
	re, err := regexp.Compile(`var trade_routes\s*=\s*({.*})`)
	if err != nil {
		return tr, err
	}
	rj := re.FindStringSubmatch(string(rd))
	if len(rj) != 2 {
		return tr, fmt.Errorf("failed to parse API response from rarethief.com")
	}

	if err := json.Unmarshal([]byte(rj[1]), &tr); err != nil {
		return tr, err
	}
	da := strings.SplitN(tr.Dates, " - ", 2)
	vf, err := time.Parse("2006/01/02", fmt.Sprintf("%v/%v", time.Now().Year(), da[0]))
	if err != nil {
		return tr, fmt.Errorf("failed to parse valid from date")
	}
	vt, err := time.Parse("2006/01/02 15:04:05", fmt.Sprintf("%v/%v 23:59:59",
		time.Now().Year(), da[1]))
	if err != nil {
		return tr, fmt.Errorf("failed to parse valid thru date")
	}
	tr.ValidFrom = vf
	tr.ValidThru = vt

	return tr, nil
}
