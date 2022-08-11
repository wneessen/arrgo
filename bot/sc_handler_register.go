package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/wneessen/arrgo/config"
	"github.com/wneessen/arrgo/crypto"
	"github.com/wneessen/arrgo/model"
)

// SlashCmdRegister handles the /register slash command
func (b *Bot) SlashCmdRegister(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	u, err := b.Model.User.GetByUserID(i.Member.User.ID)
	if err != nil && !errors.Is(err, model.ErrUserNotExistent) {
		return err
	}
	if u.ID > 0 {
		e := []*discordgo.MessageEmbed{
			{
				Type:        discordgo.EmbedTypeArticle,
				Title:       "Welcome back!",
				Description: "You are already registered with ArrGo. Thanks for double checking...",
			},
		}
		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: e}); err != nil {
			return fmt.Errorf("failed to edit /register request: %w", err)
		}
		return nil
	}

	us, err := crypto.RandomBytes(config.CryptoKeyLen)
	if err != nil {
		return fmt.Errorf("failed to generate user secret: %w", err)
	}
	ek, err := crypto.EncryptAuth(us, []byte(b.Config.Data.EncryptionKey), []byte(i.Member.User.ID))
	if err != nil {
		return fmt.Errorf("failed to encrypt user secret with global encryption key")
	}
	ui := model.User{
		UserID:        i.Member.User.ID,
		EncryptionKey: ek,
	}
	if err := b.Model.User.Insert(&ui); err != nil {
		return fmt.Errorf("failed to insert user into database: %w", err)
	}

	e := []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeArticle,
			Title:       "Welcome!",
			Description: "You have successfully registered your account and are now able to use the full feature set",
		},
	}
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: e}); err != nil {
		return fmt.Errorf("failed to edit /register request: %w", err)
	}

	return nil
}
