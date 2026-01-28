package game

// RollStats generates random starting stats for a new character.
func RollStats() Stats {
	return Stats{
		Strength: roll2d6(),
		Luck:     roll2d6(),
		Health:   roll2d6() + 6, // classic: stamina/health a bit higher
	}
}
