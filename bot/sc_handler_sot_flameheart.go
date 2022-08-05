package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/wneessen/arrgo/crypto"
	"github.com/wneessen/arrgo/model"
	"strings"
)

// SlashCmdSoTFlameheart handles the /flameheart slash command
func (b *Bot) SlashCmdSoTFlameheart(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	e, err := b.getFlameheartEmbed()
	if err != nil {
		return err
	}

	// Edit the deferred message
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: e}); err != nil {
		return err
	}

	return nil
}

// ScheduledEventSoTFlameheart performs scheuled FH spam message to the guilds system channel
func (b *Bot) ScheduledEventSoTFlameheart() error {
	ll := b.Log.With().Str("context", "bot.ScheduledEventSoTFlameheart").Logger()
	gl, err := b.Model.Guild.GetGuilds()
	if err != nil {
		return err
	}
	for _, g := range gl {
		var en bool
		en, err = b.Model.Guild.GetPrefBool(g, model.GuildPrefScheduledFlameheart)
		if err != nil {
			if !errors.Is(err, model.ErrGuildPrefNotExistant) {
				ll.Warn().Msgf("failed to read scheduled flameheart preference from DB: %s", err)
				continue
			}
			en = true
		}
		if en {
			e, err := b.getFlameheartEmbed()
			if err != nil {
				continue
			}
			if _, err := b.Session.ChannelMessageSendEmbed(g.SystemChannelID, e[0]); err != nil {
				ll.Error().Msgf("failed to send timed FH spam message: %s", err)
			}
		}
	}
	return nil
}

// getFlameheartEmbed returns a embed slice for use in slash commands or SendMessageEmbeds
func (b *Bot) getFlameheartEmbed() ([]*discordgo.MessageEmbed, error) {
	q := []string{
		`You're starting to annoy me.`,
		`Surely you don't expect to triumph?`,
		`Surely you don't expect to win?`,
		`The time for games is over!`,
		`Your supplies must be dwindling by now!`,
		`This isn't going as planned....`,
		`I'm losing my patience...`,
		`Your luck is about to run out!`,
		`I'm amazed you've survived this long.`,
		`I've been complacent, but no longer!`,
		`Surely you realise this is a lost cause?`,
		`An alliance! Against me? Ha ha ha ha…`,
		`You think you're worthy of facing me?`,
		`You sail only as long as I wish it!`,
		`I expected more resistance.`,
		`The waves are mine to command!`,
		`I'm just getting started!`,
		`You call this bravery? I call it stupidity!`,
		`Was that it?`,
		`Can you match my strength?`,
		`Don't you realise you're outnumbered?`,
		`All that you see is under my control!`,
		`You would do well to avoid me!`,
		`How much longer can you last out here?`,
		`My galleons will overwhelm you!`,
		`You're no match for me!`,
		`Let's see you handle this!`,
		`Is this your first time at the helm?`,
		`You won't last forever!`,
		`You dare defy me?!`,
		`I'll show you no mercy!`,
		`Tremble at the might of Flameheart!`,
		`\*\*\* frustrated groan ***`,
	}

	// Prepare the embed message
	rn, err := crypto.RandNum(len(q))
	if err != nil {
		return []*discordgo.MessageEmbed{}, fmt.Errorf("failed to generate random number: %w", err)
	}
	ef := []*discordgo.MessageEmbedField{
		{
			Value:  fmt.Sprintf(`«*%s*»`, strings.ToUpper(q[rn])),
			Name:   "Captain Flameheart yells at you:",
			Inline: false,
		},
	}
	e := []*discordgo.MessageEmbed{
		{
			Type: discordgo.EmbedTypeRich,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: `https://github.com/wneessen/arrgo/raw/main/assets/flameheart.png`,
			},
			Fields: ef,
		},
	}
	return e, nil
}
