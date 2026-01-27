package game

import (
	"testing"
)

func TestNewPlayer(t *testing.T) {
	start := "test_node"
	player := NewPlayer(start)

	if player.NodeID != start {
		t.Errorf("Expected NodeID %s, got %s", start, player.NodeID)
	}

	if player.Stats.Strength != 7 {
		t.Errorf("Expected default Strength 7, got %d", player.Stats.Strength)
	}

	if player.Stats.Luck != 7 {
		t.Errorf("Expected default Luck 7, got %d", player.Stats.Luck)
	}

	if player.Stats.Health != 12 {
		t.Errorf("Expected default Health 12, got %d", player.Stats.Health)
	}

	if player.Flags == nil {
		t.Error("Expected Flags map to be initialized")
	}
}

func TestCurrentNode(t *testing.T) {
	story := &Story{
		Start: "node1",
		Nodes: map[string]*Node{
			"node1": {
				Text: "Test node 1",
			},
			"node2": {
				Text: "Test node 2",
			},
		},
	}

	engine := &Engine{Story: story}
	player := NewPlayer("node1")

	node, err := engine.CurrentNode(player)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if node.Text != "Test node 1" {
		t.Errorf("Expected text 'Test node 1', got '%s'", node.Text)
	}

	// Test unknown node
	player.NodeID = "unknown"
	_, err = engine.CurrentNode(player)
	if err == nil {
		t.Error("Expected error for unknown node")
	}
}

func TestApplyChoice_Simple(t *testing.T) {
	story := &Story{
		Start: "start",
		Nodes: map[string]*Node{
			"start": {
				Text: "Start here",
				Choices: []Choice{
					{
						Key:  "next",
						Text: "Go next",
						Next: "end",
					},
				},
			},
			"end": {
				Text: "The end",
			},
		},
	}

	engine := &Engine{Story: story}
	player := NewPlayer("start")

	result, err := engine.ApplyChoice(player, "next")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.State.NodeID != "end" {
		t.Errorf("Expected NodeID 'end', got '%s'", result.State.NodeID)
	}

	if result.ErrorMessage != "" {
		t.Errorf("Expected no error message, got '%s'", result.ErrorMessage)
	}
}

func TestApplyChoice_WithEffects(t *testing.T) {
	maxHealth := 12
	story := &Story{
		Start: "start",
		Nodes: map[string]*Node{
			"start": {
				Text: "Start here",
				Choices: []Choice{
					{
						Key:  "heal",
						Text: "Heal yourself",
						Next: "start",
						Effects: []Effect{
							{
								Op:       "add",
								Stat:     "health",
								Value:    2,
								ClampMax: &maxHealth,
							},
						},
					},
				},
			},
		},
	}

	engine := &Engine{Story: story}
	player := NewPlayer("start")
	player.Stats.Health = 10

	result, err := engine.ApplyChoice(player, "heal")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.State.Stats.Health != 12 {
		t.Errorf("Expected Health 12 (clamped), got %d", result.State.Stats.Health)
	}
}

func TestApplyChoice_WithCheck(t *testing.T) {
	story := &Story{
		Start: "start",
		Nodes: map[string]*Node{
			"start": {
				Text: "Test your luck",
				Choices: []Choice{
					{
						Key:  "try",
						Text: "Try it",
						Check: &Check{
							Stat:   "luck",
							Roll:   "2d6",
							Target: "stat",
						},
						OnSuccessNext: "success",
						OnFailureNext: "failure",
					},
				},
			},
			"success": {
				Text: "You succeeded!",
			},
			"failure": {
				Text: "You failed!",
			},
		},
	}

	engine := &Engine{Story: story}
	player := NewPlayer("start")
	player.Stats.Luck = 12 // High luck, should usually succeed

	result, err := engine.ApplyChoice(player, "try")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.LastRoll == nil {
		t.Error("Expected LastRoll to be set")
	}

	if result.LastOutcome == nil {
		t.Error("Expected LastOutcome to be set")
	}

	if *result.LastRoll < 2 || *result.LastRoll > 12 {
		t.Errorf("Expected roll between 2-12, got %d", *result.LastRoll)
	}

	// With luck 12, roll should be <= 12, so should succeed
	if *result.LastOutcome != "success" && *result.LastOutcome != "failure" {
		t.Errorf("Expected outcome 'success' or 'failure', got '%s'", *result.LastOutcome)
	}
}

func TestApplyChoice_InvalidChoice(t *testing.T) {
	story := &Story{
		Start: "start",
		Nodes: map[string]*Node{
			"start": {
				Text:    "Start here",
				Choices: []Choice{},
			},
		},
	}

	engine := &Engine{Story: story}
	player := NewPlayer("start")

	result, err := engine.ApplyChoice(player, "invalid")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.ErrorMessage == "" {
		t.Error("Expected error message for invalid choice")
	}

	if result.State.NodeID != player.NodeID {
		t.Error("Expected state to remain unchanged")
	}
}

func TestApplyChoice_DestinationEffects(t *testing.T) {
	story := &Story{
		Start: "start",
		Nodes: map[string]*Node{
			"start": {
				Text: "Start here",
				Choices: []Choice{
					{
						Key:  "next",
						Text: "Go next",
						Next: "damage",
					},
				},
			},
			"damage": {
				Text: "You take damage",
				Effects: []Effect{
					{
						Op:    "add",
						Stat:  "health",
						Value: -3,
					},
				},
			},
		},
	}

	engine := &Engine{Story: story}
	player := NewPlayer("start")
	player.Stats.Health = 10

	result, err := engine.ApplyChoice(player, "next")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.State.Stats.Health != 7 {
		t.Errorf("Expected Health 7 (10 - 3), got %d", result.State.Stats.Health)
	}
}

func TestGetStat(t *testing.T) {
	player := NewPlayer("test")
	player.Stats.Strength = 10
	player.Stats.Luck = 8
	player.Stats.Health = 15

	if getStat(player, "strength") != 10 {
		t.Errorf("Expected Strength 10, got %d", getStat(player, "strength"))
	}

	if getStat(player, "luck") != 8 {
		t.Errorf("Expected Luck 8, got %d", getStat(player, "luck"))
	}

	if getStat(player, "health") != 15 {
		t.Errorf("Expected Health 15, got %d", getStat(player, "health"))
	}

	if getStat(player, "unknown") != 0 {
		t.Errorf("Expected 0 for unknown stat, got %d", getStat(player, "unknown"))
	}
}

func TestSetStat(t *testing.T) {
	player := NewPlayer("test")

	setStat(&player, "strength", 15)
	if player.Stats.Strength != 15 {
		t.Errorf("Expected Strength 15, got %d", player.Stats.Strength)
	}

	setStat(&player, "luck", 9)
	if player.Stats.Luck != 9 {
		t.Errorf("Expected Luck 9, got %d", player.Stats.Luck)
	}

	setStat(&player, "health", 20)
	if player.Stats.Health != 20 {
		t.Errorf("Expected Health 20, got %d", player.Stats.Health)
	}
}

func TestApplyEffects(t *testing.T) {
	player := NewPlayer("test")
	player.Stats.Health = 10
	maxHealth := 12
	minHealth := 1

	effects := []Effect{
		{
			Op:       "add",
			Stat:     "health",
			Value:    5,
			ClampMax: &maxHealth,
		},
		{
			Op:       "add",
			Stat:     "health",
			Value:    -20, // Would go below 0
			ClampMin: &minHealth,
		},
	}

	result := applyEffects(player, effects)

	if result.Stats.Health != 1 {
		t.Errorf("Expected Health 1 (clamped), got %d", result.Stats.Health)
	}
}

func TestCheckRoll(t *testing.T) {
	player := NewPlayer("test")
	player.Stats.Strength = 10
	player.Stats.Luck = 5

	check := Check{
		Stat:   "strength",
		Roll:   "2d6",
		Target: "stat",
	}

	// Roll of 5 should succeed (5 <= 10)
	ok, err := checkRoll(player, check, 5)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !ok {
		t.Error("Expected roll 5 to succeed against strength 10")
	}

	// Roll of 12 should fail (12 > 10)
	ok, err = checkRoll(player, check, 12)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if ok {
		t.Error("Expected roll 12 to fail against strength 10")
	}

	// Test invalid check
	invalidCheck := Check{
		Stat:   "strength",
		Roll:   "1d6",
		Target: "stat",
	}
	_, err = checkRoll(player, invalidCheck, 5)
	if err == nil {
		t.Error("Expected error for unsupported roll type")
	}
}
