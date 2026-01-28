package game

type Stats struct {
	Strength int
	Luck     int
	Health   int
}

type PlayerState struct {
	NodeID        string
	Stats         Stats
	Flags         map[string]bool
	EnemyName     string
	EnemyStrength int
	EnemyHealth   int
}

type Story struct {
	Start string           `yaml:"start"`
	Nodes map[string]*Node `yaml:"nodes"`
}

type Node struct {
	Text    string   `yaml:"text"`
	Choices []Choice `yaml:"choices"`
	Effects []Effect `yaml:"effects"`
	Ending  bool     `yaml:"ending"`
}

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

type Check struct {
	Stat   string `yaml:"stat"`   // "strength" | "luck"
	Roll   string `yaml:"roll"`   // "2d6"
	Target string `yaml:"target"` // "stat" (roll <= stat)
}

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
