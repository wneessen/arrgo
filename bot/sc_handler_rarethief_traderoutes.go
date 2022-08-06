package bot

import (
	"encoding/json"
	"fmt"
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

// ScheduledEventUpdateTradeRoutes performs scheuled updates of the TR data from rarethief.com
func (b *Bot) ScheduledEventUpdateTradeRoutes() error {
	ll := b.Log.With().Str("context", "bot.ScheduledEventUpdateTradeRoutes").Logger()
	tr, err := b.RTGetTradeRoutes()
	if err != nil {
		return fmt.Errorf("failed to fetch traderoute: %w", err)
	}
	ll.Debug().Msgf("TR: %+v", tr)
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
