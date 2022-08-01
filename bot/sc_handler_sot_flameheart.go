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
	r := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: strings.ToUpper(q[rand.Intn(len(q))]),
		},
	}
	if err := s.InteractionRespond(i.Interaction, &r); err != nil {
		return fmt.Errorf("failed to respond to /flameheart request: %w", err)
	}
	return nil
}
