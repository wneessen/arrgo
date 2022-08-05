package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/wneessen/arrgo/model"
	"strings"
)

const (
	TitleConfigUpdated = "Bot configuration updated"
)

// SlashCmdConfig handles the /config slash command
// All /config commands require admin or moderate-members permissions on the guild
func (b *Bot) SlashCmdConfig(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ll := b.Log.With().Str("context", "bot.SlashCmdConfig").Logger()
	ol := i.ApplicationCommandData().Options

	// Only admin users are allowed to execute /config commands
	r := Requester{i.Member}
	if !r.IsAdmin() && !r.CanModerateMembers() {
		ll.Warn().Msgf("non admin user tried to change configuration: %s", i.Member.User.Username)
		return fmt.Errorf("this command is only accessible for admin-user")
	}

	// Define list of config option methods
	co := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error{
		"flameheart-spam": b.configFlameheart,
	}

	// Check if provided command is available and process it
	if h, ok := co[ol[0].Name]; ok {
		if err := h(s, i); err != nil {
			return fmt.Errorf("failed to process /config %s command: %w", ol[0].Name, err)
		}
	}

	return nil
}

// configFlameheart en-/disables the Flameheart spam for a Guild
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
			Name:   TitleConfigUpdated,
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
