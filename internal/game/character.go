package game

func RollStats() Stats {
	return Stats{
		Strength: roll2d6(),
		Luck:     roll2d6(),
		Health:   roll2d6() + 6, // classic: stamina/health a bit higher
	}
}
