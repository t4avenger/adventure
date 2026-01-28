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

func TestApplyEffects_ClampStrengthAndLuckBounds(t *testing.T) {
	player := NewPlayer("test")
	player.Stats.Strength = 12
	player.Stats.Luck = 1

	effects := []Effect{
		{
			Op:    "add",
			Stat:  "strength",
			Value: 5, // would exceed MaxStat
		},
		{
			Op:    "add",
			Stat:  "luck",
			Value: -5, // would go below MinStat
		},
	}

	result := applyEffects(player, effects)

	if result.Stats.Strength != MaxStat {
		t.Errorf("Expected Strength clamped to %d, got %d", MaxStat, result.Stats.Strength)
	}
	if result.Stats.Luck != MinStat {
		t.Errorf("Expected Luck clamped to %d, got %d", MinStat, result.Stats.Luck)
	}
}

func TestHealthGameOverRoutesToDeathNode(t *testing.T) {
	story := &Story{
		Start: "start",
		Nodes: map[string]*Node{
			"start": {
				Text: "Start",
				Choices: []Choice{
					{
						Key:  "next",
						Text: "Step into danger",
						Next: "damage",
					},
				},
			},
			"damage": {
				Text: "You are gravely wounded",
				Effects: []Effect{
					{
						Op:    "add",
						Stat:  "health",
						Value: -999,
					},
				},
			},
			DeathNodeID: {
				Text:   "You have died.",
				Ending: true,
			},
		},
	}

	engine := &Engine{Story: story}
	player := NewPlayer("start")
	player.Stats.Health = 3

	result, err := engine.ApplyChoice(player, "next")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.State.NodeID != DeathNodeID {
		t.Errorf("Expected to route to death node %q, got %q", DeathNodeID, result.State.NodeID)
	}
	if result.State.Stats.Health != MinHealth {
		t.Errorf("Expected health clamped to %d, got %d", MinHealth, result.State.Stats.Health)
	}
}

func TestResolveBattleRound_HealthNeverNegative(t *testing.T) {
	story := &Story{}
	engine := &Engine{Story: story}

	player := NewPlayer("battle")
	player.Stats.Health = 1
	player.Stats.Strength = 5

	battle := Battle{
		EnemyName:     "Strong enemy",
		EnemyStrength: 10,
		EnemyHealth:   3,
		OnVictoryNext: "victory",
		OnDefeatNext:  "defeat",
	}

	result, enemyHealth, _, outcome := engine.resolveBattleRound(player, battle, battle.EnemyHealth, 1)

	if result.Stats.Health < MinHealth {
		t.Errorf("Expected health never below %d, got %d", MinHealth, result.Stats.Health)
	}
	if enemyHealth < 0 {
		t.Errorf("Expected enemy health never below 0, got %d", enemyHealth)
	}
	if outcome != "victory" && outcome != "defeat" && outcome != "player_hit" && outcome != "enemy_hit" && outcome != "tie" {
		t.Errorf("Unexpected outcome %q", outcome)
	}
}

func TestApplyChoice_BattleInitializesEnemyState(t *testing.T) {
	story := &Story{
		Start: "battle",
		Nodes: map[string]*Node{
			"battle": {
				Text: "A goblin attacks!",
				Choices: []Choice{
					{
						Key:  "attack",
						Text: "Attack",
						Mode: "battle_attack",
						Battle: &Battle{
							EnemyName:     "Goblin",
							EnemyStrength: 8,
							EnemyHealth:   3,
							OnVictoryNext: "victory",
							OnDefeatNext:  "defeat",
						},
					},
				},
			},
			"victory": {
				Text: "You won!",
			},
		},
	}

	engine := &Engine{Story: story}
	player := NewPlayer("battle")
	player.Stats.Strength = 10
	player.Stats.Health = 12

	result, err := engine.ApplyChoice(player, "attack")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Enemy state should be initialized
	if result.State.EnemyName != "Goblin" {
		t.Errorf("Expected EnemyName 'Goblin', got '%s'", result.State.EnemyName)
	}
	if result.State.EnemyStrength != 8 {
		t.Errorf("Expected EnemyStrength 8, got %d", result.State.EnemyStrength)
	}
	if result.State.EnemyHealth <= 0 {
		t.Errorf("Expected EnemyHealth > 0, got %d", result.State.EnemyHealth)
	}
}

func TestApplyChoice_BattleClearsEnemyStateOnVictory(t *testing.T) {
	story := &Story{
		Start: "battle",
		Nodes: map[string]*Node{
			"battle": {
				Text: "A weak enemy",
				Choices: []Choice{
					{
						Key:  "attack",
						Text: "Attack",
						Mode: "battle_attack",
						Battle: &Battle{
							EnemyName:     "Weakling",
							EnemyStrength: 1,
							EnemyHealth:   1,
							OnVictoryNext: "victory",
						},
					},
				},
			},
			"victory": {
				Text: "You won!",
			},
		},
	}

	engine := &Engine{Story: story}
	player := NewPlayer("battle")
	player.Stats.Strength = 12
	player.Stats.Health = 12

	// Set enemy state first
	player.EnemyName = "Weakling"
	player.EnemyStrength = 1
	player.EnemyHealth = 1

	// Run multiple rounds until victory (may take a few tries)
	for i := 0; i < 10; i++ {
		result, err := engine.ApplyChoice(player, "attack")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		player = result.State

		if result.State.NodeID == "victory" {
			// Enemy state should be cleared on victory
			if result.State.EnemyHealth != 0 {
				t.Errorf("Expected EnemyHealth 0 on victory, got %d", result.State.EnemyHealth)
			}
			if result.State.EnemyName != "" {
				t.Errorf("Expected EnemyName empty on victory, got '%s'", result.State.EnemyName)
			}
			if result.State.EnemyStrength != 0 {
				t.Errorf("Expected EnemyStrength 0 on victory, got %d", result.State.EnemyStrength)
			}
			return
		}
	}

	t.Error("Battle did not resolve to victory after 10 rounds")
}

func TestApplyChoice_RunAwayClearsEnemyState(t *testing.T) {
	story := &Story{
		Start: "battle",
		Nodes: map[string]*Node{
			"battle": {
				Text: "A goblin blocks your path",
				Choices: []Choice{
					{
						Key:  "attack",
						Text: "Attack",
						Mode: "battle_attack",
						Battle: &Battle{
							EnemyName:     "Goblin",
							EnemyStrength: 8,
							EnemyHealth:   3,
						},
					},
					{
						Key:  "run",
						Text: "Run away",
						Next: "safe",
					},
				},
			},
			"safe": {
				Text: "You escaped",
			},
		},
	}

	engine := &Engine{Story: story}
	player := NewPlayer("battle")
	player.EnemyName = "Goblin"
	player.EnemyStrength = 8
	player.EnemyHealth = 2

	result, err := engine.ApplyChoice(player, "run")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Enemy state should be cleared when running away
	if result.State.EnemyHealth != 0 {
		t.Errorf("Expected EnemyHealth 0 after running, got %d", result.State.EnemyHealth)
	}
	if result.State.EnemyName != "" {
		t.Errorf("Expected EnemyName empty after running, got '%s'", result.State.EnemyName)
	}
	if result.State.EnemyStrength != 0 {
		t.Errorf("Expected EnemyStrength 0 after running, got %d", result.State.EnemyStrength)
	}
	if result.State.NodeID != "safe" {
		t.Errorf("Expected NodeID 'safe', got '%s'", result.State.NodeID)
	}
}

func TestApplyChoice_LuckAttackReducesLuck(t *testing.T) {
	story := &Story{
		Start: "battle",
		Nodes: map[string]*Node{
			"battle": {
				Text: "Battle",
				Choices: []Choice{
					{
						Key:  "luck",
						Text: "Luck attack",
						Mode: "battle_luck",
						Battle: &Battle{
							EnemyName:     "Enemy",
							EnemyStrength: 5,
							EnemyHealth:   3,
						},
					},
				},
			},
		},
	}

	engine := &Engine{Story: story}
	player := NewPlayer("battle")
	player.Stats.Luck = 7
	player.Stats.Strength = 10
	player.Stats.Health = 12

	result, err := engine.ApplyChoice(player, "luck")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Luck should be reduced by 1 (but clamped to minimum 1)
	if result.State.Stats.Luck != 6 {
		t.Errorf("Expected Luck 6 (7-1), got %d", result.State.Stats.Luck)
	}

	// Test that luck doesn't go below 1
	player.Stats.Luck = 1
	result, err = engine.ApplyChoice(player, "luck")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.State.Stats.Luck < MinStat {
		t.Errorf("Expected Luck >= %d, got %d", MinStat, result.State.Stats.Luck)
	}
}
