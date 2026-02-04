// Package game provides the core game engine for text-based adventure games.
// It handles story loading, player state management, choice resolution,
// stat checks, and combat mechanics.
package game

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"unicode"
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

	// HordeName is the display name when 4+ enemies are combined.
	HordeName = "Horde"
)

// getBattleEnemies returns initial enemy state from battle (Enemies list or legacy single-enemy fields).
func getBattleEnemies(b *Battle) []EnemyState {
	if len(b.Enemies) > 0 {
		out := make([]EnemyState, 0, len(b.Enemies))
		for _, e := range b.Enemies {
			h := e.Health
			if h <= 0 {
				h = 1
			}
			out = append(out, EnemyState{Name: e.Name, Strength: e.Strength, Health: h})
		}
		return out
	}
	if b.EnemyName != "" || b.EnemyHealth > 0 {
		h := b.EnemyHealth
		if h <= 0 {
			h = 1
		}
		return []EnemyState{{Name: b.EnemyName, Strength: b.EnemyStrength, Health: h}}
	}
	return nil
}

// collapseToHorde returns a single "Horde" entry if len(es) > 3.
func collapseToHorde(es []EnemyState) []EnemyState {
	if len(es) <= 3 {
		return es
	}
	sumHealth := 0
	sumStr := 0
	for _, e := range es {
		sumHealth += e.Health
		sumStr += e.Strength
	}
	meanStr := sumStr / len(es)
	if meanStr < MinStat {
		meanStr = MinStat
	}
	return []EnemyState{{Name: HordeName, Strength: meanStr, Health: sumHealth}}
}

// DefaultStoryID is the story ID used for new sessions when no choice has been made.
const DefaultStoryID = "demo"

// Engine manages game state and resolves player choices.
type Engine struct {
	Stories map[string]*Story // story ID -> Story
}

// StepResult contains the result of applying a player choice, including
// the updated state, any dice rolls, and outcome messages.
type StepResult struct {
	State          PlayerState
	LastRoll       *int
	LastPlayerDice *[2]int // two d6 values for display
	LastEnemyDice  *[2]int // battle only
	LastOutcome    *string // "success"/"failure"
	ErrorMessage   string
}

// DefaultAvatar is the avatar ID used for new players.
const DefaultAvatar = "male_young"

// NewPlayer creates a new player state with default starting stats for the given story.
func NewPlayer(storyID, startNodeID string) PlayerState {
	return PlayerState{
		NodeID:  startNodeID,
		StoryID: storyID,
		Name:    "",
		Avatar:  DefaultAvatar,
		Stats: Stats{
			Strength: 7,
			Luck:     7,
			Health:   12,
		},
		Flags:        map[string]bool{},
		VisitedNodes: []string{startNodeID},
	}
}

// story returns the story for the player's StoryID; if missing or empty, uses default. Returns nil if no story found.
func (e *Engine) story(st *PlayerState) *Story {
	id := st.StoryID
	if id == "" {
		id = DefaultStoryID
	}
	return e.Stories[id]
}

// CurrentNode returns the node the player is currently on.
func (e *Engine) CurrentNode(st *PlayerState) (*Node, error) {
	s := e.story(st)
	if s == nil {
		return nil, fmt.Errorf("unknown story: %s", st.StoryID)
	}
	n := s.Nodes[st.NodeID]
	if n == nil {
		return nil, fmt.Errorf("unknown node: %s", st.NodeID)
	}
	return n, nil
}

// ApplyChoice processes a player's choice, updating their state and
// determining the next node in the story.
func (e *Engine) ApplyChoice(st *PlayerState, choiceKey string) (StepResult, error) {
	return e.ApplyChoiceWithAnswer(st, choiceKey, "")
}

// ApplyChoiceWithAnswer processes a player's choice and optional typed answer,
// updating their state and determining the next node in the story.
func (e *Engine) ApplyChoiceWithAnswer(st *PlayerState, choiceKey, answer string) (StepResult, error) {
	node, err := e.CurrentNode(st)
	if err != nil {
		return StepResult{}, err
	}

	var ch *Choice
	for i := range node.Choices {
		if node.Choices[i].Key == choiceKey {
			ch = &node.Choices[i]
			break
		}
	}
	// Dynamic battle keys: "battle:attack:0", "battle:luck:1", "battle:run"
	if ch == nil {
		for i := range node.Choices {
			c := &node.Choices[i]
			if c.Battle != nil && (choiceKey == c.Key || strings.HasPrefix(choiceKey, c.Key+":")) {
				ch = c
				break
			}
		}
	}
	if ch == nil {
		return StepResult{State: *st, ErrorMessage: "That choice doesn't exist."}, nil
	}

	var lastRoll *int
	var lastPlayerDice *[2]int
	var lastEnemyDice *[2]int
	var lastOutcome *string

	next := ch.Next
	if ch.Prompt != nil {
		promptNext, promptMsg := resolvePromptNext(ch.Prompt, answer, ch.Next)
		if promptNext == "" {
			return StepResult{State: *st, ErrorMessage: promptMsg}, nil
		}
		next = promptNext
		applyEffects(st, ch.Effects)
	} else {
		// Apply node-level effects first (optional; here we only do choice effects + destination effects)
		applyEffects(st, ch.Effects)
	}
	if ch.Check != nil && ch.Prompt == nil {
		d1, d2 := roll2d6Ex()
		roll := d1 + d2
		lastRoll = &roll
		lastPlayerDice = &[2]int{d1, d2}

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

	// Battle: multi-enemy (Enemies list) or legacy single enemy.
	if ch.Battle != nil && ch.Prompt == nil {
		battleNext := e.applyBattle(st, ch, choiceKey, &lastRoll, &lastOutcome, &lastPlayerDice, &lastEnemyDice)
		if battleNext != "" {
			next = battleNext
		}
	} else if len(st.Enemies) > 0 {
		// Non-battle choice while in combat (e.g. run from another choice): clear enemies.
		st.Enemies = nil
	}

	if next == "" {
		return StepResult{State: *st, ErrorMessage: "No destination for that choice."}, nil
	}

	oldNodeID := st.NodeID
	st.NodeID = next
	if st.VisitedNodes == nil {
		st.VisitedNodes = []string{}
	}
	if st.NodeID != oldNodeID {
		st.VisitedNodes = append(st.VisitedNodes, st.NodeID)
	}

	// Apply destination node effects on entry, but avoid re-applying the same
	// node's effects when we intentionally stay on the same node (e.g. during
	// multi-round battles).
	s := e.story(st)
	if s != nil && st.NodeID != oldNodeID {
		dst := s.Nodes[st.NodeID]
		if dst != nil && len(dst.Effects) > 0 {
			applyEffects(st, dst.Effects)
		}
	}

	// Global health-based game over: if health is 0 or below after all
	// effects, transition to a dedicated death node when available.
	if st.Stats.Health <= MinHealth {
		st.Stats.Health = MinHealth
		if s != nil {
			if _, ok := s.Nodes[DeathNodeID]; ok {
				if st.NodeID != DeathNodeID {
					st.VisitedNodes = append(st.VisitedNodes, DeathNodeID)
				}
				st.NodeID = DeathNodeID
			}
		}
	}

	return StepResult{State: *st, LastRoll: lastRoll, LastPlayerDice: lastPlayerDice, LastEnemyDice: lastEnemyDice, LastOutcome: lastOutcome}, nil
}

// applyBattle handles one battle round (or run). Returns next node ID or "" if caller should keep next.
func (e *Engine) applyBattle(st *PlayerState, ch *Choice, choiceKey string, lastRoll **int, lastOutcome **string, lastPlayerDice, lastEnemyDice **[2]int) string {
	b := ch.Battle
	// Initialize enemies from battle if first round.
	if len(st.Enemies) == 0 {
		st.Enemies = collapseToHorde(getBattleEnemies(b))
		if len(st.Enemies) == 0 {
			return b.OnVictoryNext
		}
	}

	// Parse action: "run", "attack:N", "luck:N" or legacy exact key (attack:0 / luck:0 from ch.Mode).
	var action string
	var enemyIndex int
	if strings.HasPrefix(choiceKey, ch.Key+":") {
		action = choiceKey[len(ch.Key)+1:]
	} else {
		// Legacy single-enemy choice: treat as attack:0 or luck:0 from mode.
		if ch.Mode == "battle_luck" {
			action = "luck:0"
		} else {
			action = "attack:0"
		}
	}

	if action == "run" {
		st.Enemies = nil
		return ch.Next
	}

	// Parse "attack:N" or "luck:N"
	isLuck := strings.HasPrefix(action, "luck:")
	if !isLuck && !strings.HasPrefix(action, "attack:") {
		return ""
	}
	idxStr := action[strings.Index(action, ":")+1:]
	n, err := strconv.Atoi(idxStr)
	if err != nil || n < 0 || n >= len(st.Enemies) {
		return ""
	}
	enemyIndex = n

	playerDamage := 1
	if isLuck {
		st.Stats.Luck--
		if st.Stats.Luck < MinStat {
			st.Stats.Luck = MinStat
		}
		playerDamage = 2
	}

	enemyStr := st.Enemies[enemyIndex].Strength
	enemyHp := st.Enemies[enemyIndex].Health
	updatedSt, newHealth, playerDice, enemyDice, outcome := e.resolveBattleRound(st, enemyStr, enemyHp, playerDamage)
	*st = *updatedSt
	if playerDice != nil {
		pd := *playerDice
		ed := *enemyDice
		*lastPlayerDice = &pd
		*lastEnemyDice = &ed
		sum := pd[0] + pd[1]
		*lastRoll = &sum
	}
	if outcome != "" {
		*lastOutcome = &outcome
	}

	st.Enemies[enemyIndex].Health = newHealth
	if newHealth <= 0 {
		st.Enemies = append(st.Enemies[:enemyIndex], st.Enemies[enemyIndex+1:]...)
	}
	if len(st.Enemies) == 0 {
		if b.OnVictoryNext != "" {
			return b.OnVictoryNext
		}
		return ""
	}
	if outcome == OutcomeDefeat {
		st.Enemies = nil
		return DeathNodeID
	}
	return st.NodeID
}

// resolveBattleRound runs a single opposed-roll round between the player and
// one enemy (strength + health). Returns updated player state, new enemy health, player/enemy dice, outcome.
func (e *Engine) resolveBattleRound(st *PlayerState, enemyStrength, enemyHealth, playerDamage int) (updatedState *PlayerState, newEnemyHealth int, playerDice, enemyDice *[2]int, outcome string) {
	if enemyHealth <= 0 {
		enemyHealth = 1
	}

	pd1, pd2 := roll2d6Ex()
	ed1, ed2 := roll2d6Ex()
	playerRoll := pd1 + pd2
	enemyRoll := ed1 + ed2

	playerTotal := st.Stats.Strength + playerRoll
	enemyTotal := enemyStrength + enemyRoll

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
	playerDice = &[2]int{pd1, pd2}
	enemyDice = &[2]int{ed1, ed2}
	return updatedState, newEnemyHealth, playerDice, enemyDice, outcome
}

// HasEnemies returns true if the player is in an active battle.
func (st *PlayerState) HasEnemies() bool {
	return len(st.Enemies) > 0
}

func resolvePromptNext(prompt *Prompt, answer, fallback string) (next, message string) {
	if prompt == nil {
		if fallback != "" {
			return fallback, ""
		}
		return "", "No destination for that choice."
	}

	normalized := normalizeAnswer(answer)
	if normalized == "" {
		return "", promptFailureMessage(prompt, "Please enter an answer.")
	}

	for _, ans := range prompt.Answers {
		if ans.Next == "" {
			continue
		}
		if promptAnswerMatches(ans, normalized) {
			return ans.Next, ""
		}
	}

	if prompt.DefaultNext != "" {
		return prompt.DefaultNext, ""
	}
	if fallback != "" {
		return fallback, ""
	}
	return "", promptFailureMessage(prompt, "That does not seem right.")
}

func promptAnswerMatches(ans Answer, normalized string) bool {
	if ans.Match != "" {
		if normalizeAnswer(ans.Match) == normalized {
			return true
		}
	}
	for _, m := range ans.Matches {
		if normalizeAnswer(m) == normalized {
			return true
		}
	}
	return false
}

func promptFailureMessage(prompt *Prompt, fallback string) string {
	if prompt != nil && prompt.FailureMessage != "" {
		return prompt.FailureMessage
	}
	return fallback
}

func normalizeAnswer(answer string) string {
	answer = strings.ToLower(answer)
	var b strings.Builder
	for _, r := range answer {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
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

// roll2d6Ex returns the two d6 values for display.
func roll2d6Ex() (d1, d2 int) {
	d1 = d6()
	d2 = d6()
	return d1, d2
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
