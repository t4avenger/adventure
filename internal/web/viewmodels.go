package web

import "adventure/internal/game"

// StartViewModel contains data for rendering the character creation screen.
type StartViewModel struct {
	Stats        game.Stats
	StrengthDice [2]int // two d6 for Strength
	LuckDice     [2]int
	HealthDice   [2]int
	SessionID    string // so Begin request can use same session if cookie not sent
}
