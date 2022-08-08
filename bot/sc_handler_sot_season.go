package bot

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math"
)

// SoTSeasonList represents the JSON structure of the Sea of Thieves seasons API response
type SoTSeasonList []SoTSeasonProgress

// SoTSeasonProgress represents the JSON structure of the Sea of Thieves seasons progress within a season API response
type SoTSeasonProgress struct {
	LevelProgress       float64         `json:"LevelProgress"`
	Tier                int             `json:"Tier"`
	SeasonTitle         string          `json:"Title"`
	TotalChallenges     int             `json:"TotalChallenges"`
	CompletedChallenges int             `json:"CompleteChallenges"`
	Tiers               []SoTSeasonTier `json:"Tiers"`
	CDNPath             string          `json:"CdnPath"`
}

// SoTSeasonTier represents the JSON structure of the Sea of Thieves seasons tier within the
// season progress API response
type SoTSeasonTier struct {
	Number int              `json:"Number"`
	Title  string           `json:"Title"`
	Levels []SoTSeasonLevel `json:"Levels"`
}

// SoTSeasonLevel represents the JSON structure of the Sea of Thieves seasons levels within the
// season progress API response
type SoTSeasonLevel struct {
	Number  int              `json:"Number"`
	Rewards SoTSeasonRewards `json:"RewardsV2"`
}

// SoTSeasonRewards represents the JSON structure of the Sea of Thieves seasons rewards collection
// within the season progress API response
type SoTSeasonRewards struct {
	Base       []SoTSeasonReward `json:"Base"`
	Legendary  []SoTSeasonReward `json:"Legendary"`
	SeasonPass []SoTSeasonReward `json:"SeasonPass"`
}

// SoTSeasonReward represents the JSON structure of the Sea of Thieves seasons reward in the
// level within the season progress API response
type SoTSeasonReward struct {
	CurrencyType           string `json:"CurrencyType"`
	Locked                 bool   `json:"Locked"`
	Owned                  bool   `json:"Owned"`
	EntitlementURL         string `json:"EntitlementUrl"`
	EntitlementText        string `json:"EntitlementText"`
	EntitlementDescription string `json:"EntitlementDescription"`
}

// SlashCmdSoTSeasonProgress handles the /season slash command
func (b *Bot) SlashCmdSoTSeasonProgress(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	r := &Requester{i.Member, b.Model.User, nil}
	sl, err := b.SoTGetSeasonProgress(r)
	if err != nil {
		return err
	}
	if len(sl) <= 0 {
		return fmt.Errorf("no SoT season progress found")
	}

	// Roman numerals map
	rn := map[int]string{
		1:  "I",
		2:  "II",
		3:  "III",
		4:  "IV",
		5:  "V",
		6:  "VI",
		7:  "VII",
		8:  "VIII",
		9:  "IX",
		10: "X",
	}

	sp := sl[len(sl)-1]
	var ef []*discordgo.MessageEmbedField
	ef = append(ef, &discordgo.MessageEmbedField{
		Name:   "Current title",
		Value:  sp.Tiers[sp.Tier-1].Title,
		Inline: false,
	})
	ef = append(ef, &discordgo.MessageEmbedField{
		Name:   "Renown Level",
		Value:  fmt.Sprintf("🌡️ %.1f%%", sp.LevelProgress),
		Inline: true,
	})
	ef = append(ef, &discordgo.MessageEmbedField{
		Name:   "Renown Tier",
		Value:  fmt.Sprintf("📜 %d", sp.Tier),
		Inline: true,
	})
	ef = append(ef, &discordgo.MessageEmbedField{
		Name:   "Challenges",
		Value:  fmt.Sprintf("☑️ %d/%d completed", sp.CompletedChallenges, sp.TotalChallenges),
		Inline: true,
	})

	e := []*discordgo.MessageEmbed{
		{
			Title: fmt.Sprintf("Your progress in Sea of Thieves %s", sp.SeasonTitle),
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: fmt.Sprintf("https://github.com/wneessen/arrgo/raw/main/assets/numerals/%s.png",
					rn[sp.Tier]),
			},
			Type:   discordgo.EmbedTypeRich,
			Fields: ef,
		},
	}
	l := sp.Tiers[sp.Tier-1].Levels
	if len(l) > 0 {
		pl := int(math.Floor(sp.LevelProgress))
		for _, cl := range l {
			if cl.Number == pl {
				if len(cl.Rewards.Base) > 0 {
					br := cl.Rewards.Base[0]
					e = append(e, buildSoTRewardEmbed("Base", &br, sp.CDNPath))
				}
				if len(cl.Rewards.Legendary) > 0 {
					br := cl.Rewards.Legendary[0]
					e = append(e, buildSoTRewardEmbed("Legendary", &br, sp.CDNPath))
				}
				if len(cl.Rewards.SeasonPass) > 0 {
					br := cl.Rewards.SeasonPass[0]
					e = append(e, buildSoTRewardEmbed("Season Pass", &br, sp.CDNPath))
				}
			}
		}
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: e}); err != nil {
		return err
	}
	return nil
}

// SoTGetSeasonProgress returns the parsed API response from the Sea of Thieves season progress API
func (b *Bot) SoTGetSeasonProgress(rq *Requester) (SoTSeasonList, error) {
	var s SoTSeasonList
	hc, err := NewHTTPClient()
	if err != nil {
		return s, fmt.Errorf(ErrFailedHTTPClient, err)
	}
	c, err := rq.GetSoTRATCookie()
	if err != nil {
		return s, err
	}
	r, err := hc.HttpReq(ApiURLSoTSeasons, ReqMethodGet, nil)
	if err != nil {
		return s, err
	}
	r.SetSOTRequest(c)
	rd, _, err := hc.Fetch(r)
	if err != nil {
		return s, err
	}
	if err := json.Unmarshal(rd, &s); err != nil {
		return s, err
	}
	return s, nil
}

// buildRewardEmbed returns a discordgo.MessageEmbed object for different reward types
func buildSoTRewardEmbed(t string, r *SoTSeasonReward, cp string) *discordgo.MessageEmbed {
	e := &discordgo.MessageEmbed{
		Title: fmt.Sprintf("Your latest reward in the %q tier", t),
		Type:  discordgo.EmbedTypeImage,
	}
	switch r.CurrencyType {
	case "gold-s":
		e.Description = "A nice stack of Gold!"
		e.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: "https://github.com/wneessen/arrgo/raw/main/assets/season/gold-s.png",
		}
	case "doubloons-s":
		e.Description = "A nice stack of Doubloons!"
		e.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: "https://github.com/wneessen/arrgo/raw/main/assets/season/doubloons-s.png",
		}
	default:
		e.Description = r.EntitlementText
		e.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: fmt.Sprintf("%s/%s", cp, r.EntitlementURL),
		}
	}
	return e
}
