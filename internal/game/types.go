package game

// Stats represents a character's core attributes.
type Stats struct {
	Strength int
	Luck     int
	Health   int
}

// PlayerState tracks the current game state for a player, including
// their location, stats, flags, and any active enemy in combat.
type PlayerState struct {
	NodeID        string
	Stats         Stats
	Flags         map[string]bool
	EnemyName     string
	EnemyStrength int
	EnemyHealth   int
}

// Story represents a complete adventure story with nodes and choices.
type Story struct {
	Start string           `yaml:"start"`
	Nodes map[string]*Node `yaml:"nodes"`
}

// Node represents a single location or scene in the adventure.
type Node struct {
	Text    string   `yaml:"text"`
	Choices []Choice `yaml:"choices"`
	Effects []Effect `yaml:"effects"`
	Ending  bool     `yaml:"ending"`
}

// Choice represents a player action available at a node.
type Choice struct {
	Key           string   `yaml:"key"`
	Text          string   `yaml:"text"`
	Next          string   `yaml:"next"`
	Mode          string   `yaml:"mode"` // e.g. "battle_attack", "battle_luck"
	Check         *Check   `yaml:"check"`
	OnSuccessNext string   `yaml:"onSuccessNext"`
	OnFailureNext string   `yaml:"onFailureNext"`
	Effects       []Effect `yaml:"effects"`
	Battle        *Battle  `yaml:"battle"`
}

// Check defines a stat check that must be passed to proceed.
type Check struct {
	Stat   string `yaml:"stat"`   // "strength" | "luck"
	Roll   string `yaml:"roll"`   // "2d6"
	Target string `yaml:"target"` // "stat" (roll <= stat)
}

// Effect modifies player stats when applied.
type Effect struct {
	Op       string `yaml:"op"`   // "add"
	Stat     string `yaml:"stat"` // "health" | "strength" | "luck"
	Value    int    `yaml:"value"`
	ClampMax *int   `yaml:"clampMax"`
	ClampMin *int   `yaml:"clampMin"`
}

// Battle describes an opposed-roll combat where both player and enemy
// roll 2d6 and add their Strength. The higher total scores a hit and
// deals damage to the other side's Health. The engine resolves one
// round at a time; the story can keep the player on the same node for
// multiple rounds or branch on victory/defeat.
type Battle struct {
	EnemyName     string `yaml:"enemyName"`
	EnemyStrength int    `yaml:"enemyStrength"`
	EnemyHealth   int    `yaml:"enemyHealth"`

	OnVictoryNext string `yaml:"onVictoryNext"`
	OnDefeatNext  string `yaml:"onDefeatNext"`
}
