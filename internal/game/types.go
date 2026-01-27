package game

type Stats struct {
	Strength int
	Luck     int
	Health   int
}

type PlayerState struct {
	NodeID string
	Stats  Stats
	Flags  map[string]bool
}

type Story struct {
	Start string           `yaml:"start"`
	Nodes map[string]*Node `yaml:"nodes"`
}

type Node struct {
	Text    string    `yaml:"text"`
	Choices []Choice  `yaml:"choices"`
	Effects []Effect  `yaml:"effects"`
	Ending  bool      `yaml:"ending"`
}

type Choice struct {
	Key            string  `yaml:"key"`
	Text           string  `yaml:"text"`
	Next           string  `yaml:"next"`
	Check          *Check  `yaml:"check"`
	OnSuccessNext  string  `yaml:"onSuccessNext"`
	OnFailureNext  string  `yaml:"onFailureNext"`
	Effects        []Effect `yaml:"effects"`
}

type Check struct {
	Stat   string `yaml:"stat"`   // "strength" | "luck"
	Roll   string `yaml:"roll"`   // "2d6"
	Target string `yaml:"target"` // "stat" (roll <= stat)
}

type Effect struct {
	Op       string `yaml:"op"`       // "add"
	Stat     string `yaml:"stat"`     // "health" | "strength" | "luck"
	Value    int    `yaml:"value"`
	ClampMax *int   `yaml:"clampMax"`
	ClampMin *int   `yaml:"clampMin"`
}
