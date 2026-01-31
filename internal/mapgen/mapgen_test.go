package mapgen

import (
	"testing"

	"adventure/internal/game"
)

func TestGenerate_NilStory(t *testing.T) {
	b, err := Generate(nil, []string{"a"}, "a", "Test")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if b != nil {
		t.Error("expected nil PDF for nil story")
	}
}

func TestGenerate_EmptyVisitedUsesCurrent(t *testing.T) {
	st := &game.Story{
		Start: "a",
		Nodes: map[string]*game.Node{
			"a": {Text: "Start", Scenery: "forest"},
		},
	}
	b, err := Generate(st, nil, "a", "Test")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(b) < 100 {
		t.Errorf("PDF too short: %d bytes", len(b))
	}
	if !bytesPrefix(b, []byte("%PDF")) {
		t.Error("output is not a PDF (missing %PDF header)")
	}
}

func TestGenerate_PathReturnsPDF(t *testing.T) {
	st := &game.Story{
		Start: "a",
		Title: "Test",
		Nodes: map[string]*game.Node{
			"a": {Text: "Start", Scenery: "shore", Choices: []game.Choice{{Key: "n", Next: "b"}}},
			"b": {Text: "Forest", Scenery: "forest", Choices: []game.Choice{{Key: "n", Next: "c"}}},
			"c": {Text: "Bridge", Scenery: "bridge", Ending: true},
		},
	}
	visited := []string{"a", "b", "c"}
	b, err := Generate(st, visited, "b", "Test Adventure")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(b) < 100 {
		t.Errorf("PDF too short: %d bytes", len(b))
	}
	if !bytesPrefix(b, []byte("%PDF")) {
		t.Error("output is not a PDF (missing %PDF header)")
	}
}

func bytesPrefix(b, prefix []byte) bool {
	if len(b) < len(prefix) {
		return false
	}
	for i := range prefix {
		if b[i] != prefix[i] {
			return false
		}
	}
	return true
}
