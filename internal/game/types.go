package game

// Stats represents a character's core attributes.
type Stats struct {
	Strength int
	Luck     int
	Health   int
}

// EnemyState represents one enemy in combat (current health etc.).
type EnemyState struct {
	Name     string
	Strength int
	Health   int
}

// PlayerState tracks the current game state for a player, including
// their location, stats, flags, and any active enemies in combat.
type PlayerState struct {
	NodeID       string
	StoryID      string // adventure ID e.g. "demo"
	Name         string // character display name
	Avatar       string // avatar ID e.g. "male_young"
	Stats        Stats
	RerollUsed   bool // true once stats have been rerolled on setup
	Flags        map[string]bool
	Enemies      []EnemyState // 1–3 shown individually; 4+ stored as one "Horde" entry
	VisitedNodes []string     // node IDs in order visited (for treasure map)
}

// Story represents a complete adventure story with nodes and choices.
type Story struct {
	Title string           `yaml:"title"` // optional display name; if empty, derived from ID
	Start string           `yaml:"start"`
	Nodes map[string]*Node `yaml:"nodes"`
}

// Node represents a single location or scene in the adventure.
type Node struct {
	Text           string   `yaml:"text"`
	Scenery        string   `yaml:"scenery"`         // scenery image filename (with or without extension) in story's scenery/ dir e.g. "forest", "forest.png"; empty = default
	Audio          string   `yaml:"audio"`           // optional audio filename (with or without extension) in story's audio/ dir e.g. "forest_ambient"; empty = none
	EntryAnimation string   `yaml:"entry_animation"` // e.g. "door_open"; empty = none
	Choices        []Choice `yaml:"choices"`
	Effects        []Effect `yaml:"effects"`
	Ending         bool     `yaml:"ending"`
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
	Prompt        *Prompt  `yaml:"prompt"`
}

// Prompt defines a question that expects a typed answer.
// Answers route to different nodes; DefaultNext is used when no match is found.
type Prompt struct {
	Question       string   `yaml:"question"`
	Placeholder    string   `yaml:"placeholder"`
	Answers        []Answer `yaml:"answers"`
	DefaultNext    string   `yaml:"defaultNext"`
	FailureMessage string   `yaml:"failureMessage"`
}

// Answer maps one or more expected strings to a destination node.
type Answer struct {
	Match   string   `yaml:"match"`
	Matches []string `yaml:"matches"`
	Next    string   `yaml:"next"`
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

// Enemy is a single enemy definition in story YAML.
type Enemy struct {
	Name     string `yaml:"name"`
	Strength int    `yaml:"strength"`
	Health   int    `yaml:"health"`
}

// Battle describes an opposed-roll combat where both player and enemy
// roll 2d6 and add their Strength. The higher total scores a hit and
// deals damage to the other side's Health. The engine resolves one
// round at a time; the story can keep the player on the same node for
// multiple rounds or branch on victory/defeat.
// Use Enemies for multiple foes; legacy single-enemy fields are used when Enemies is empty.
type Battle struct {
	Enemies []Enemy `yaml:"enemies"`

	EnemyName     string `yaml:"enemyName"`
	EnemyStrength int    `yaml:"enemyStrength"`
	EnemyHealth   int    `yaml:"enemyHealth"`

	OnVictoryNext string `yaml:"onVictoryNext"`
	OnDefeatNext  string `yaml:"onDefeatNext"`
}
