package web

import "adventure/internal/game"

// AvatarOptions is the list of allowed avatar IDs for validation and templates.
var AvatarOptions = []string{"male_young", "male_old", "female_young", "female_old"}

// AdventureOption is one selectable adventure (ID and display name).
type AdventureOption struct {
	ID   string
	Name string
}

// StartViewModel contains data for rendering the character creation screen.
type StartViewModel struct {
	Stats            game.Stats
	StrengthDice     [2]int // two d6 for Strength
	LuckDice         [2]int
	HealthDice       [2]int
	RerollUsed       bool
	SessionID        string   // so Begin request can use same session if cookie not sent
	Name             string   // character display name
	Avatar           string   // avatar ID
	AvatarOptions    []string // allowed avatar IDs for the selector
	StoryID          string   // selected adventure ID
	AdventureOptions []AdventureOption
}
