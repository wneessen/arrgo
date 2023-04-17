package bot

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/wneessen/arrgo/model"
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
	r, err := b.NewRequester(i.Interaction)
	if err != nil {
		return err
	}
	if !r.IsAdmin() && !r.CanModerateMembers() {
		ll.Warn().Msgf("non admin user tried to change configuration: %s", i.Member.User.Username)
		return fmt.Errorf("this command is only accessible for admin-user")
	}

	// Define list of config option methods
	co := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error{
		"flameheart-spam":      b.configFlameheart,
		"announce-sot-summary": b.configAnnounceSoTPlaySummary,
		"announce-channel":     b.overrideAnnounceChannel,
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
	nv, err := appCommandGetEnalbedDisabled(i.ApplicationCommandData().Options)
	if err != nil {
		return err
	}

	g, err := b.Model.Guild.GetByGuildID(i.GuildID)
	if err != nil {
		return fmt.Errorf(ErrFailedGuildLookupDB, err)
	}
	if err = b.Model.Guild.SetPref(g, model.GuildPrefScheduledFlameheart, nv); err != nil {
		return fmt.Errorf("failed to set flameheart preference in database: %w", err)
	}

	e := []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeArticle,
			Title:       TitleConfigUpdated,
			Description: "The bot will not spam the server with Captain Flameheart quotes",
		},
	}
	if nv {
		e[0].Description = "The bot will spam the server with Captain Flameheart quotes"
	}

	// Edit the deferred message
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &e}); err != nil {
		return fmt.Errorf("failed to edit /config flameheart-spam request: %w", err)
	}

	return nil
}

// overrideAnnounceChannel overrides the default system channel with a guild specific channel
func (b *Bot) overrideAnnounceChannel(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	mo := i.ApplicationCommandData().Options
	if len(mo) <= 0 {
		return fmt.Errorf("no options found")
	}
	rc, ok := mo[0].Value.(string)
	if !ok {
		return fmt.Errorf("provided value is not a string")
	}

	re, err := regexp.Compile(`<#(\d+)>`)
	if err != nil {
		return err
	}
	cha := re.FindStringSubmatch(rc)
	if len(cha) != 2 {
		return fmt.Errorf("failed to parse value string")
	}
	ch := cha[1]
	g, err := b.Model.Guild.GetByGuildID(i.GuildID)
	if err != nil {
		return fmt.Errorf(ErrFailedGuildLookupDB, err)
	}
	if err := b.Model.Guild.SetPref(g, model.GuildPrefAnnounceChannel, ch); err != nil {
		return err
	}

	// Edit the deferred message
	e := []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeArticle,
			Title:       TitleConfigUpdated,
			Description: fmt.Sprintf("The annoucment channel for this server has been set to: <#%s>", ch),
		},
	}
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &e}); err != nil {
		return fmt.Errorf("failed to edit /override annouce-channel request: %w", err)
	}

	return nil
}

// configAnnounceSoTPlaySummary en-/disables the announcing of SoT play summaries
func (b *Bot) configAnnounceSoTPlaySummary(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	nv, err := appCommandGetEnalbedDisabled(i.ApplicationCommandData().Options)
	if err != nil {
		return err
	}

	g, err := b.Model.Guild.GetByGuildID(i.GuildID)
	if err != nil {
		return fmt.Errorf(ErrFailedGuildLookupDB, err)
	}
	if err = b.Model.Guild.SetPref(g, model.GuildPrefAnnounceSoTSummary, nv); err != nil {
		return fmt.Errorf("failed to set announce-sot-summary preference in database: %w", err)
	}

	e := []*discordgo.MessageEmbed{
		{
			Type:        discordgo.EmbedTypeArticle,
			Title:       TitleConfigUpdated,
			Description: "The bot will not announce a user's summary after they played SoT",
		},
	}
	if nv {
		e[0].Description = "The bot will announce a user's summary after they played SoT"
	}

	// Edit the deferred message
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &e}); err != nil {
		return fmt.Errorf("failed to edit /config announce-sot-summary request: %w", err)
	}

	return nil
}

// getEnabledDisabled takes the applicationcommand options and checks wether enabled or disabled was selected
func appCommandGetEnalbedDisabled(os []*discordgo.ApplicationCommandInteractionDataOption) (bool, error) {
	if len(os) <= 0 {
		return false, fmt.Errorf("no options found")
	}
	so := os[0].Options
	if len(so) <= 0 {
		return false, fmt.Errorf("no suboption found")
	}

	var nv bool
	switch strings.ToLower(so[0].Name) {
	case "disable":
		nv = false
	case "enable":
		nv = true
	default:
		return false, fmt.Errorf("unsupported value")
	}
	return nv, nil
}
