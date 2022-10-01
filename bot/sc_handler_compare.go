package bot

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"time"
)

// SlashCmdSoTCompare handles the /compare slash command
func (b *Bot) SlashCmdSoTCompare(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ol := i.ApplicationCommandData().Options
	if len(ol) <= 0 {
		return fmt.Errorf("no duration given")
	}
	ds, ok := ol[0].Value.(string)
	if !ok {
		return fmt.Errorf("failed to cast provided hour value as string")
	}
	d, err := time.ParseDuration(fmt.Sprintf("-%sh", ds))
	if err != nil {
		return err
	}
	ots := time.Now().Add(d)

	u, err := b.Model.User.GetByUserID(i.Member.User.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve user from DB: %w", err)
	}
	if err := b.StoreSoTUserStats(u); err != nil {
		return fmt.Errorf("failed to update user stats in DB: %w", err)
	}
	cus, err := b.Model.UserStats.GetByUserID(u.ID)
	if err != nil {
		return err
	}
	ous, err := b.Model.UserStats.GetByUserIDAtTime(u.ID, ots)
	if err != nil {
		return err
	}

	p := message.NewPrinter(language.German)
	var ef []*discordgo.MessageEmbedField
	if cus.Gold != ous.Gold {
		v := cus.Gold - ous.Gold
		ef = append(ef, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Gold", IconGold),
			Value:  fmt.Sprintf("%s **%s** Gold", changeIcon(v), p.Sprintf("%d", v)),
			Inline: true,
		})
	}
	if cus.Doubloons != ous.Doubloons {
		v := cus.Doubloons - ous.Doubloons
		ef = append(ef, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Doubloons", IconDoubloon),
			Value:  fmt.Sprintf("%s **%s** Doubloons", changeIcon(v), p.Sprintf("%d", v)),
			Inline: true,
		})
	}
	if cus.AncientCoins != ous.AncientCoins {
		v := cus.AncientCoins - ous.AncientCoins
		ef = append(ef, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Ancient Coins", IconAncientCoin),
			Value:  fmt.Sprintf("%s **%s** Ancient Coins", changeIcon(v), p.Sprintf("%d", v)),
			Inline: true,
		})
	}
	if cus.KrakenDefeated != ous.KrakenDefeated {
		v := cus.KrakenDefeated - ous.KrakenDefeated
		ef = append(ef, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Kraken", IconKraken),
			Value:  fmt.Sprintf("**%s** defeated", p.Sprintf("%d", v)),
			Inline: true,
		})
	}
	if cus.MegalodonEnounter != ous.MegalodonEnounter {
		v := cus.MegalodonEnounter - ous.MegalodonEnounter
		ef = append(ef, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Megalodon", IconMegalodon),
			Value:  fmt.Sprintf("**%s** encounter(s)", p.Sprintf("%d", v)),
			Inline: true,
		})
	}
	if cus.ChestsHandedIn != ous.ChestsHandedIn {
		v := cus.ChestsHandedIn - ous.ChestsHandedIn
		ef = append(ef, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Chests", IconChest),
			Value:  fmt.Sprintf("**%s** handed in", p.Sprintf("%d", v)),
			Inline: true,
		})
	}
	if cus.ShipsSunk != ous.ShipsSunk {
		v := cus.ShipsSunk - ous.ShipsSunk
		ef = append(ef, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Other Ships", IconShip),
			Value:  fmt.Sprintf("**%s** sunk", p.Sprintf("%d", v)),
			Inline: true,
		})
	}
	if cus.VomittedTimes != ous.VomittedTimes {
		v := cus.VomittedTimes - ous.VomittedTimes
		ef = append(ef, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Vomitted", IconVomit),
			Value:  fmt.Sprintf("**%s** times", p.Sprintf("%d", v)),
			Inline: true,
		})
	}
	if cus.DistanceSailed != ous.DistanceSailed {
		v := cus.DistanceSailed - ous.DistanceSailed
		ef = append(ef, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s Distance", IconDistance),
			Value:  fmt.Sprintf("**%s** nmi sailed", p.Sprintf("%d", v/1852)),
			Inline: true,
		})
	}
	for len(ef)%3 != 0 {
		ef = append(ef, &discordgo.MessageEmbedField{
			Value:  "\U0000FEFF",
			Name:   "\U0000FEFF",
			Inline: true,
		})
	}

	var e []*discordgo.MessageEmbed
	if len(ef) > 0 {
		e = []*discordgo.MessageEmbed{
			{
				Title: fmt.Sprintf("Your user statistics overview compared to %s hours ago in Sea of Thieves",
					ds),
				Type:   discordgo.EmbedTypeRich,
				Fields: ef,
			},
		}
	} else {
		e = []*discordgo.MessageEmbed{
			{
				Title: fmt.Sprintf("None of your users stats in Sea of Thieves changed within the last "+
					"%s hours", ds),
				Type:   discordgo.EmbedTypeRich,
				Fields: ef,
			},
		}
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Embeds: &e}); err != nil {
		return err
	}
	return nil
}
