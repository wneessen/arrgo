package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"strings"
)

// SlashCmdSoTFlameheart handles the /flameheart slash command
func (b *Bot) SlashCmdSoTFlameheart(s *discordgo.Session, i *discordgo.InteractionCreate) error {
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
	}

	// Initalize the deferred message
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: ""},
	})
	if err != nil {
		return fmt.Errorf("failed to defer /flameheart request: %w", err)
	}

	// Prepare the embed message
	ef := []*discordgo.MessageEmbedField{
		{
			Value:  strings.ToUpper(q[rand.Intn(len(q))]),
			Name:   "Captain Flameheart yells at you!",
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

	// Edit the deferred message
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: e}); err != nil {
		return fmt.Errorf("failed to edit /flameheart request: %w", err)
	}

	return nil
}
