// Package game provides the core game engine for text-based adventure games.
// It handles story loading, player state management, choice resolution,
// stat checks, and combat mechanics.
package game

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

const (
	// MinStat and MaxStat bound Strength and Luck.
	MinStat = 1
	// MaxStat is the maximum value for Strength and Luck stats.
	MaxStat = 12

	// MinHealth is the lowest health a player can have; 0 means dead.
	MinHealth = 0

	// DeathNodeID is the special node the story can define to represent
	// a generic death/game-over screen.
	DeathNodeID = "death"

	// OutcomeSuccess indicates a successful stat check roll.
	OutcomeSuccess = "success"
	// OutcomeFailure indicates a failed stat check roll.
	OutcomeFailure = "failure"
	// OutcomeVictory indicates the player won a battle.
	OutcomeVictory = "victory"
	// OutcomeDefeat indicates the player lost a battle.
	OutcomeDefeat = "defeat"
	// OutcomeTie indicates a battle round ended in a tie.
	OutcomeTie = "tie"
	// OutcomePlayerHit indicates the player hit the enemy in battle.
	OutcomePlayerHit = "player_hit"
	// OutcomeEnemyHit indicates the enemy hit the player in battle.
	OutcomeEnemyHit = "enemy_hit"

	// StatStrength is the stat name for strength.
	StatStrength = "strength"
	// StatLuck is the stat name for luck.
	StatLuck = "luck"
	// StatHealth is the stat name for health.
	StatHealth = "health"

	// OpAdd is the effect operation for adding to a stat.
	OpAdd = "add"
)

// Engine manages game state and resolves player choices.
type Engine struct {
	Story *Story
}

// StepResult contains the result of applying a player choice, including
// the updated state, any dice rolls, and outcome messages.
type StepResult struct {
	State        PlayerState
	LastRoll     *int
	LastOutcome  *string // "success"/"failure"
	ErrorMessage string
}

// NewPlayer creates a new player state with default starting stats.
func NewPlayer(start string) PlayerState {
	return PlayerState{
		NodeID: start,
		Stats: Stats{
			Strength: 7,
			Luck:     7,
			Health:   12,
		},
		Flags: map[string]bool{},
	}
}

// CurrentNode returns the node the player is currently on.
func (e *Engine) CurrentNode(st *PlayerState) (*Node, error) {
	n := e.Story.Nodes[st.NodeID]
	if n == nil {
		return nil, fmt.Errorf("unknown node: %s", st.NodeID)
	}
	return n, nil
}

// ApplyChoice processes a player's choice, updating their state and
// determining the next node in the story.
func (e *Engine) ApplyChoice(st *PlayerState, choiceKey string) (StepResult, error) {
	node, err := e.CurrentNode(st)
	if err != nil {
		return StepResult{}, err
	}

	var ch *Choice
	// Linear search is acceptable here as nodes typically have < 10 choices
	for i := range node.Choices {
		if node.Choices[i].Key == choiceKey {
			ch = &node.Choices[i]
			break
		}
	}
	if ch == nil {
		return StepResult{State: *st, ErrorMessage: "That choice doesn't exist."}, nil
	}

	// Apply node-level effects first (optional; here we only do choice effects + destination effects)
	applyEffects(st, ch.Effects)

	var lastRoll *int
	var lastOutcome *string

	next := ch.Next
	if ch.Check != nil {
		roll := roll2d6()
		lastRoll = &roll

		ok, err := checkRoll(st, *ch.Check, roll)
		if err != nil {
			return StepResult{State: *st, ErrorMessage: err.Error()}, nil
		}
		var outcome string
		if ok {
			outcome = OutcomeSuccess
		} else {
			outcome = OutcomeFailure
		}
		lastOutcome = &outcome

		if ok && ch.OnSuccessNext != "" {
			next = ch.OnSuccessNext
		}
		if !ok && ch.OnFailureNext != "" {
			next = ch.OnFailureNext
		}
	}

	// Battle (opposed Strength + 2d6 rolls for player and enemy), resolved
	// one round at a time so the player can choose actions each round.
	if ch.Battle != nil {
		// Initialize enemy state if this is the first round.
		if st.EnemyHealth <= 0 {
			st.EnemyHealth = ch.Battle.EnemyHealth
			if st.EnemyHealth <= 0 {
				st.EnemyHealth = 1
			}
			st.EnemyName = ch.Battle.EnemyName
			st.EnemyStrength = ch.Battle.EnemyStrength
		}

		playerDamage := 1
		// Luck-based attack: spend 1 Luck (clamped) and deal extra damage
		// to the enemy on a successful hit.
		if ch.Mode == "battle_luck" {
			st.Stats.Luck--
			if st.Stats.Luck < MinStat {
				st.Stats.Luck = MinStat
			}
			playerDamage = 2
		}

		var battleLastRoll *int
		var battleOutcome string
		var updatedSt *PlayerState
		var newEnemyHealth int
		updatedSt, newEnemyHealth, battleLastRoll, battleOutcome = e.resolveBattleRound(st, *ch.Battle, st.EnemyHealth, playerDamage)
		*st = *updatedSt
		st.EnemyHealth = newEnemyHealth

		if battleLastRoll != nil {
			lastRoll = battleLastRoll
		}
		if battleOutcome != "" {
			lastOutcome = &battleOutcome
		}

		switch battleOutcome {
		case OutcomeVictory:
			// Clear enemy state when battle is won
			st.EnemyHealth = 0
			st.EnemyName = ""
			st.EnemyStrength = 0
			if ch.Battle.OnVictoryNext != "" {
				next = ch.Battle.OnVictoryNext
			}
		case OutcomeDefeat:
			// Clear enemy state when player dies
			st.EnemyHealth = 0
			st.EnemyName = ""
			st.EnemyStrength = 0
			next = DeathNodeID
		default:
			// Continue fighting on the same node; let the UI render another
			// round of choices.
			next = st.NodeID
		}
	} else if st.EnemyHealth > 0 {
		// If this choice doesn't have a battle, clear enemy state (e.g., running away)
		st.EnemyHealth = 0
		st.EnemyName = ""
		st.EnemyStrength = 0
	}

	if next == "" {
		return StepResult{State: *st, ErrorMessage: "No destination for that choice."}, nil
	}

	oldNodeID := st.NodeID
	st.NodeID = next

	// Apply destination node effects on entry, but avoid re-applying the same
	// node's effects when we intentionally stay on the same node (e.g. during
	// multi-round battles).
	if st.NodeID != oldNodeID {
		dst := e.Story.Nodes[st.NodeID]
		if dst != nil && len(dst.Effects) > 0 {
			applyEffects(st, dst.Effects)
		}
	}

	// Global health-based game over: if health is 0 or below after all
	// effects, transition to a dedicated death node when available.
	if st.Stats.Health <= MinHealth {
		st.Stats.Health = MinHealth
		if _, ok := e.Story.Nodes[DeathNodeID]; ok {
			st.NodeID = DeathNodeID
		}
	}

	return StepResult{State: *st, LastRoll: lastRoll, LastOutcome: lastOutcome}, nil
}

// resolveBattleRound runs a single opposed-roll round between the player and
// the configured enemy. It returns the updated player state, the new enemy
// health, the player's roll for the round, and an outcome string:
//   - "player_hit" (enemy took damage but survived)
//   - "enemy_hit"  (player took damage but survived)
//   - "tie"        (no damage dealt)
//   - "victory"    (enemy defeated)
//   - "defeat"     (player reduced to 0 health)
func (e *Engine) resolveBattleRound(st *PlayerState, b Battle, enemyHealth, playerDamage int) (updatedState *PlayerState, newEnemyHealth int, rollResult *int, outcome string) {
	if enemyHealth <= 0 {
		enemyHealth = 1
	}

	playerRoll := roll2d6()
	enemyRoll := roll2d6()

	playerTotal := st.Stats.Strength + playerRoll
	enemyTotal := b.EnemyStrength + enemyRoll

	outcome = OutcomeTie

	// Create a copy to avoid mutating the input
	result := *st

	switch {
	case playerTotal > enemyTotal:
		enemyHealth -= playerDamage
		if enemyHealth <= 0 {
			enemyHealth = 0
			outcome = OutcomeVictory
		} else {
			outcome = OutcomePlayerHit
		}
	case enemyTotal > playerTotal:
		result.Stats.Health--
		if result.Stats.Health <= MinHealth {
			result.Stats.Health = MinHealth
			outcome = OutcomeDefeat
		} else {
			outcome = OutcomeEnemyHit
		}
		// default case: outcome already set to OutcomeTie above
	}

	updatedState = &result
	newEnemyHealth = enemyHealth
	rollResult = &playerRoll
	return updatedState, newEnemyHealth, rollResult, outcome
}

func checkRoll(st *PlayerState, c Check, roll int) (bool, error) {
	if c.Roll != "2d6" || c.Target != "stat" {
		return false, fmt.Errorf("unsupported check: roll=%s target=%s", c.Roll, c.Target)
	}
	stat := getStat(st, c.Stat)
	return roll <= stat, nil
}

func getStat(st *PlayerState, stat string) int {
	switch stat {
	case StatStrength:
		return st.Stats.Strength
	case StatLuck:
		return st.Stats.Luck
	case StatHealth:
		return st.Stats.Health
	default:
		return 0
	}
}

func setStat(st *PlayerState, stat string, v int) {
	switch stat {
	case StatStrength:
		st.Stats.Strength = v
	case StatLuck:
		st.Stats.Luck = v
	case StatHealth:
		st.Stats.Health = v
	}
}

func applyEffects(st *PlayerState, effs []Effect) {
	for _, ef := range effs {
		if ef.Op != OpAdd {
			continue
		}
		cur := getStat(st, ef.Stat)
		nv := cur + ef.Value

		if ef.ClampMax != nil && nv > *ef.ClampMax {
			nv = *ef.ClampMax
		}
		if ef.ClampMin != nil && nv < *ef.ClampMin {
			nv = *ef.ClampMin
		}

		// Apply global bounds for stats regardless of story-provided
		// clamps so that rules are always enforced.
		switch ef.Stat {
		case StatStrength, StatLuck:
			if nv < MinStat {
				nv = MinStat
			}
			if nv > MaxStat {
				nv = MaxStat
			}
		case StatHealth:
			if nv < MinHealth {
				nv = MinHealth
			}
		}

		setStat(st, ef.Stat, nv)
	}
}

// crypto-rand small helper; plenty for a adventure
func roll2d6() int {
	return d6() + d6()
}

func d6() int {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		// Fallback to a simple pseudo-random if crypto/rand fails
		// This should never happen in practice, but we handle it gracefully
		return 1
	}
	n := binary.LittleEndian.Uint64(b[:])
	// n%6 is safe: result is 0-5, adding 1 gives 1-6
	return int(n%6) + 1 //nolint:gosec // modulo 6 is safe, result fits in int
}
