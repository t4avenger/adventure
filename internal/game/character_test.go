package game

import (
	"testing"
)

func TestRollStats(t *testing.T) {
	// Roll stats multiple times to ensure they're in expected ranges
	for i := 0; i < 100; i++ {
		stats := RollStats()

		// Strength: 2d6 = 2-12
		if stats.Strength < 2 || stats.Strength > 12 {
			t.Errorf("Strength out of range: got %d, expected 2-12", stats.Strength)
		}

		// Luck: 2d6 = 2-12
		if stats.Luck < 2 || stats.Luck > 12 {
			t.Errorf("Luck out of range: got %d, expected 2-12", stats.Luck)
		}

		// Health: 2d6 + 6 = 8-18
		if stats.Health < 8 || stats.Health > 18 {
			t.Errorf("Health out of range: got %d, expected 8-18", stats.Health)
		}
	}
}

func TestRollStats_Distribution(t *testing.T) {
	// Test that we get a reasonable distribution of values
	strengthCounts := make(map[int]int)
	luckCounts := make(map[int]int)
	healthCounts := make(map[int]int)

	iterations := 1000
	for i := 0; i < iterations; i++ {
		stats := RollStats()
		strengthCounts[stats.Strength]++
		luckCounts[stats.Luck]++
		healthCounts[stats.Health]++
	}

	// Check that we get multiple different values (not all the same)
	if len(strengthCounts) < 5 {
		t.Errorf("Expected at least 5 different strength values, got %d", len(strengthCounts))
	}

	if len(luckCounts) < 5 {
		t.Errorf("Expected at least 5 different luck values, got %d", len(luckCounts))
	}

	if len(healthCounts) < 5 {
		t.Errorf("Expected at least 5 different health values, got %d", len(healthCounts))
	}
}

func TestRollStatsDetailed(t *testing.T) {
	for i := 0; i < 50; i++ {
		stats, dice := RollStatsDetailed()

		if stats.Strength != dice[0][0]+dice[0][1] {
			t.Errorf("Strength %d != dice sum %d+%d", stats.Strength, dice[0][0], dice[0][1])
		}
		if stats.Luck != dice[1][0]+dice[1][1] {
			t.Errorf("Luck %d != dice sum %d+%d", stats.Luck, dice[1][0], dice[1][1])
		}
		if stats.Health != dice[2][0]+dice[2][1]+6 {
			t.Errorf("Health %d != dice sum %d+%d+6", stats.Health, dice[2][0], dice[2][1])
		}
		for j := 0; j < 3; j++ {
			for k := 0; k < 2; k++ {
				if dice[j][k] < 1 || dice[j][k] > 6 {
					t.Errorf("dice[%d][%d] = %d, expected 1-6", j, k, dice[j][k])
				}
			}
		}
	}
}
