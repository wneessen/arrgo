package bot

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
)

// changeIcon returns either an increase or decrease icon based on the provided value
func changeIcon[V int | int64 | float32 | float64](v V) string {
	if v > 0 {
		return IconIncrease
	}
	return IconDecrease
}
