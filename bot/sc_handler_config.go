package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/wneessen/arrgo/model"
	"strings"
)

const (
	TitleConfigUpdates = "Bot configuration updated"
)

// SlashCmdConfig handles the /config slash command
func (b *Bot) SlashCmdConfig(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ll := b.Log.With().Str("context", "bot.SlashCmdConfig").Logger()
	ol := i.ApplicationCommandData().Options

	// Initalize the deferred message
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "",
			Flags:   uint64(discordgo.MessageFlagsEphemeral),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to defer /config request: %w", err)
	}

	// Define list of config option methods
	co := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error{
		"flameheart-spam": b.configFlameheart,
	}

	// Check if provided command is available and process it
	if h, ok := co[ol[0].Name]; ok {
		if err := h(s, i); err != nil {
			ll.Error().Msgf("failed to process /config %s command: %s", ol[0].Name, err)
		}
	}

	return nil
}

func (b *Bot) configFlameheart(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	mo := i.ApplicationCommandData().Options
	if len(mo) <= 0 {
		return fmt.Errorf("no options found")
	}
	so := mo[0].Options
	if len(so) <= 0 {
		return fmt.Errorf("no suboption found")
	}

	g, err := b.Model.Guild.GetByGuildID(i.GuildID)
	if err != nil {
		return fmt.Errorf("failed to look up guild in database: %w", err)
	}

	var nv bool
	switch strings.ToLower(so[0].Name) {
	case "disable":
		nv = false
	case "enable":
		nv = true
	default:
		return fmt.Errorf("unsupported value")
	}
	if err = b.Model.Guild.SetPref(g, model.GuildPrefScheduledFlameheart, nv); err != nil {
		return fmt.Errorf("failed to set flameheart preference in database: %w", err)
	}

	// Initalize the deferred message
	ef := []*discordgo.MessageEmbedField{
		{
			Value:  "The bot will not spam the server with Captain Flameheart quotes",
			Name:   TitleConfigUpdates,
			Inline: false,
		},
	}
	if nv {
		ef[0].Value = "The bot will spam the server with Captain Flameheart quotes"
	}

	e := []*discordgo.MessageEmbed{
		{
			Type:   discordgo.EmbedTypeRich,
			Fields: ef,
		},
	}

	// Edit the deferred message
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: e}); err != nil {
		return fmt.Errorf("failed to edit /config flameheart-spam request: %w", err)
	}

	return nil
}
