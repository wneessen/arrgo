package bot

import "strings"

// List of icons/emojis
const (
	IconGold        = "\U0001F7E1"
	IconDoubloon    = "🔵"
	IconAncientCoin = "💰"
	IconIncrease    = "📈 "
	IconDecrease    = "📉 "
	IconArrowUp     = "⬆️ "
	IconArrowDown   = "⬇️ "
	IconKraken      = "🐙"
	IconMegalodon   = "🦈"
	IconChest       = "🗝️"
	IconShip        = "⛵"
	IconVomit       = "🤮"
	IconDistance    = "📐"
	IconGauge       = "🌡️"
	IconDuration    = "⏱️"
)

// changeIcon returns either an increase or decrease icon based on the provided value
func changeIcon[V int | int64 | float32 | float64](v V) string {
	if v > 0 {
		return IconIncrease
	}
	return IconDecrease
}

// dbEmissaryToName converts the emissary name in the DB to the human readable format
func dbEmissaryToName(e string) string {
	// factiong|hunterscall|merchantalliance|bilgerats|talltales|athenasfortune|` +
	//		`goldhoarders|orderofsouls|reapersbones
	switch strings.ToLower(e) {
	case "factiong":
		return "Guardians of Fortune"
	case "factionb":
		return "Servants of the Flame"
	case "hunterscall":
		return "Hunter's Call"
	case "merchantalliance":
		return "Merchant Alliance"
	case "bilgerats":
		return "Bilge Rats"
	case "talltales":
		return "Tall Tales"
	case "athenasfortune":
		return "Athena's Forutne"
	case "goldhoarders":
		return "Gold Hoarders"
	case "orderofsouls":
		return "Order of Souls"
	case "reapersbones":
		return "Reaper's Bones"
	default:
		return ""
	}
}
