package game

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

type Engine struct {
	Story *Story
}

type StepResult struct {
	State        PlayerState
	LastRoll     *int
	LastOutcome  *string // "success"/"failure"
	ErrorMessage string
}

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

func (e *Engine) CurrentNode(st PlayerState) (*Node, error) {
	n := e.Story.Nodes[st.NodeID]
	if n == nil {
		return nil, fmt.Errorf("unknown node: %s", st.NodeID)
	}
	return n, nil
}

func (e *Engine) ApplyChoice(st PlayerState, choiceKey string) (StepResult, error) {
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
	if ch == nil {
		return StepResult{State: st, ErrorMessage: "That choice doesn't exist."}, nil
	}

	// Apply node-level effects first (optional; here we only do choice effects + destination effects)
	st = applyEffects(st, ch.Effects)

	var lastRoll *int
	var lastOutcome *string

	next := ch.Next
	if ch.Check != nil {
		roll := roll2d6()
		lastRoll = &roll

		ok, err := checkRoll(st, *ch.Check, roll)
		if err != nil {
			return StepResult{State: st, ErrorMessage: err.Error()}, nil
		}
		out := "failure"
		if ok {
			out = "success"
		}
		lastOutcome = &out

		if ok && ch.OnSuccessNext != "" {
			next = ch.OnSuccessNext
		}
		if !ok && ch.OnFailureNext != "" {
			next = ch.OnFailureNext
		}
	}

	if next == "" {
		return StepResult{State: st, ErrorMessage: "No destination for that choice."}, nil
	}

	st.NodeID = next

	// Apply destination node effects on entry
	dst := e.Story.Nodes[st.NodeID]
	if dst != nil && len(dst.Effects) > 0 {
		st = applyEffects(st, dst.Effects)
	}

	return StepResult{State: st, LastRoll: lastRoll, LastOutcome: lastOutcome}, nil
}

func checkRoll(st PlayerState, c Check, roll int) (bool, error) {
	if c.Roll != "2d6" || c.Target != "stat" {
		return false, fmt.Errorf("unsupported check: roll=%s target=%s", c.Roll, c.Target)
	}
	stat := getStat(st, c.Stat)
	return roll <= stat, nil
}

func getStat(st PlayerState, stat string) int {
	switch stat {
	case "strength":
		return st.Stats.Strength
	case "luck":
		return st.Stats.Luck
	case "health":
		return st.Stats.Health
	default:
		return 0
	}
}

func setStat(st *PlayerState, stat string, v int) {
	switch stat {
	case "strength":
		st.Stats.Strength = v
	case "luck":
		st.Stats.Luck = v
	case "health":
		st.Stats.Health = v
	}
}

func applyEffects(st PlayerState, effs []Effect) PlayerState {
	for _, ef := range effs {
		if ef.Op != "add" {
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
		setStat(&st, ef.Stat, nv)
	}
	return st
}

// crypto-rand small helper; plenty for a adventure
func roll2d6() int {
	return d6() + d6()
}

func d6() int {
	var b [8]byte
	_, _ = rand.Read(b[:])
	n := binary.LittleEndian.Uint64(b[:])
	return int(n%6) + 1
}
