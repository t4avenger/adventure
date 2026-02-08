package game

// RollStats generates random starting stats for a new character.
func RollStats() Stats {
	stats, _ := RollStatsDetailed()
	return stats
}

// RollStatsDetailed returns stats and the two d6 values used for each (Strength, Luck, Health).
// Strength and Health use 2d6+6 so the dice pairs are the two d6 before the +6.
func RollStatsDetailed() (stats Stats, dice [3][2]int) {
	s1, s2 := roll2d6Ex()
	l1, l2 := roll2d6Ex()
	h1, h2 := roll2d6Ex()
	stats = Stats{
		Strength: s1 + s2 + 6,
		Luck:     l1 + l2,
		Health:   h1 + h2 + 6, // classic: stamina/health a bit higher
	}
	dice = [3][2]int{{s1, s2}, {l1, l2}, {h1, h2}}
	return stats, dice
}
