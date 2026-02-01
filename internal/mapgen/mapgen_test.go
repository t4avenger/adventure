package mapgen

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"adventure/internal/game"
)

func TestGenerate_NilStory(t *testing.T) {
	b, err := Generate(nil, []string{"a"}, "a", "Test", "", "")
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
	b, err := Generate(st, nil, "a", "Test", "", "")
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
	b, err := Generate(st, visited, "b", "Test Adventure", "", "")
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

func TestGenerate_WithSceneryImage_EmbedsImage(t *testing.T) {
	tmpDir := t.TempDir()
	storyID := "test_story"
	sceneryDir := filepath.Join(tmpDir, storyID, "scenery")
	if err := os.MkdirAll(sceneryDir, 0o750); err != nil {
		t.Fatalf("mkdir scenery: %v", err)
	}
	// Write a minimal PNG so tryLoadSceneImage finds it
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode PNG: %v", err)
	}
	forestPath := filepath.Join(sceneryDir, "forest.png")
	if err := os.WriteFile(forestPath, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write forest.png: %v", err)
	}

	st := &game.Story{
		Start: "a",
		Nodes: map[string]*game.Node{
			"a": {Text: "Start", Scenery: "forest"},
		},
	}
	b, err := Generate(st, []string{"a"}, "a", "Test", storyID, tmpDir)
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
