package bot

import (
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
)

// SoTAchievementList represents the JSON structure of the Sea of Thieves achievements API response
type SoTAchievementList struct {
	Sorted []SoTSortedAchievement `json:"sorted"`
}

// SoTSortedAchievement is a subpart of the Sea of Thieves achievements API response
type SoTSortedAchievement struct {
	Achievement SoTAchievement `json:"achievement"`
}

// SoTAchievement is a single achievement in the Sea of Thieves achievements API response
type SoTAchievement struct {
	Sort        int    `json:"Sort"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	MediaUrl    string `json:"MediaUrl"`
}

// SlashCmdSoTAchievement handles the /achievement slash command
func (b *Bot) SlashCmdSoTAchievement(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	r := &Requester{i.Member, b.Model.User}
	al, err := b.SoTGetAchievements(r)
	if err != nil {
		return err
	}
	if len(al.Sorted) <= 0 {
		return fmt.Errorf("no SoT achievements found")
	}

	a := al.Sorted[0].Achievement
	e := []*discordgo.MessageEmbed{
		{
			Title:       fmt.Sprintf("Your latest Sea of Thieves achievement: %s", a.Name),
			Description: a.Description,
			Image: &discordgo.MessageEmbedImage{
				URL: a.MediaUrl,
			},
			Type: discordgo.EmbedTypeImage,
		},
	}
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: e}); err != nil {
		return err
	}
	return nil
}

// SoTGetAchievements returns the parsed API response from the Sea of Thieves achievements API
func (b *Bot) SoTGetAchievements(rq *Requester) (SoTAchievementList, error) {
	var a SoTAchievementList
	c, err := rq.GetSoTRATCookie()
	if err != nil {
		return a, err
	}
	r, err := b.HTTPClient.HttpReq(ApiURLSoTAchievements, ReqMethodGet, nil)
	if err != nil {
		return a, err
	}
	r.SetSOTRequest(c)
	rd, _, err := b.HTTPClient.Fetch(r)
	if err != nil {
		return a, err
	}
	if err := json.Unmarshal(rd, &a); err != nil {
		return a, err
	}
	return a, nil
}
