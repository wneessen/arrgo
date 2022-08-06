package bot

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/wneessen/arrgo/model"
)

// SoTRATCookie represents the JSON formated Sea of Thieves authentication cookie
type SoTRATCookie struct {
	Value      string `json:"Value"`
	Expiration int64  `json:"Expiration"`
}

// SlashCmdSetRAT handles the /setrat slash command
func (b *Bot) SlashCmdSetRAT(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ol := i.ApplicationCommandData().Options

	u, err := b.Model.User.GetByUserID(i.Member.User.ID)
	if err != nil {
		if !errors.Is(err, model.ErrUserNotExistant) {
			return fmt.Errorf("failed to look up user: %w", err)
		}
		e := []*discordgo.MessageEmbed{
			{
				Type:  discordgo.EmbedTypeArticle,
				Title: "Please register your user first!",
				Description: "To use the Sea of Thieves bot features, please first use the **/register** " +
					"command to register your user with the bot",
			},
		}
		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: e}); err != nil {
			return err
		}
		return nil
	}

	if len(ol) <= 0 {
		return fmt.Errorf("no RAT cookie provided")
	}
	on := ol[0].Name
	if on != "rat-cookie" {
		return fmt.Errorf("provided option is not a rat-cookie")
	}
	ov, ok := ol[0].Value.(string)
	if !ok {
		return fmt.Errorf("unable to cast rat-cookie value as string")
	}
	if len(ov) <= 0 {
		return fmt.Errorf("provided rat-cookie cannot be empty")
	}

	// Base64-decode and JSON unmarshall the cookie
	var src SoTRATCookie
	rc, err := base64.StdEncoding.DecodeString(ov)
	if err != nil {
		return fmt.Errorf("failed to base64 decode RAT cookie: %w", err)
	}
	if err := json.Unmarshal(rc, &src); err != nil {
		return fmt.Errorf("failed to JSON unmarshall RAT cookie: %w", err)
	}

	if err := b.Model.User.SetPrefEnc(u, model.UserPrefSoTAuthToken, src.Value); err != nil {
		return fmt.Errorf("failed to store RAT cookie in DB: %w", err)
	}
	if err := b.Model.User.SetPrefEnc(u, model.UserPrefSoTAuthTokenExpiration, src.Expiration); err != nil {
		return fmt.Errorf("failed to store RAT cookie expiration date in DB: %w", err)
	}

	e := []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeArticle,
			Title:       "Sea of Thieves authentication cookie stored/updated",
			Description: "Thank you for storing/updating your Sea of Thieves authentication cookie.",
		},
	}
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: e}); err != nil {
		return err
	}
	return nil
}
