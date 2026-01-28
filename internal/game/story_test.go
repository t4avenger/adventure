package game

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadStory_Valid(t *testing.T) {
	// Create a temporary YAML file
	tmpDir := t.TempDir()
	storyPath := filepath.Join(tmpDir, "test_story.yaml")

	storyYAML := `start: "node1"
nodes:
  node1:
    text: "First node"
    choices:
      - key: "next"
        text: "Go to next"
        next: "node2"
  node2:
    text: "Second node"
    ending: true
`

	err := os.WriteFile(storyPath, []byte(storyYAML), 0o600) //nolint:gosec // test file permissions are acceptable
	if err != nil {
		t.Fatalf("Failed to create test story file: %v", err)
	}

	story, err := LoadStory(storyPath)
	if err != nil {
		t.Fatalf("Unexpected error loading story: %v", err)
	}

	if story.Start != "node1" {
		t.Errorf("Expected start node 'node1', got '%s'", story.Start)
	}

	if story.Nodes == nil {
		t.Fatal("Expected nodes map to be initialized")
	}

	node1, ok := story.Nodes["node1"]
	if !ok {
		t.Fatal("Expected node1 to exist")
	}

	if node1.Text != "First node" {
		t.Errorf("Expected node1 text 'First node', got '%s'", node1.Text)
	}

	if len(node1.Choices) != 1 {
		t.Errorf("Expected 1 choice in node1, got %d", len(node1.Choices))
	}

	if node1.Choices[0].Key != "next" {
		t.Errorf("Expected choice key 'next', got '%s'", node1.Choices[0].Key)
	}

	node2, ok := story.Nodes["node2"]
	if !ok {
		t.Fatal("Expected node2 to exist")
	}

	if !node2.Ending {
		t.Error("Expected node2 to be an ending")
	}
}

func TestLoadStory_InvalidFile(t *testing.T) {
	_, err := LoadStory("non_existent_file.yaml")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestLoadStory_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	storyPath := filepath.Join(tmpDir, "invalid.yaml")

	invalidYAML := `start: "node1"
nodes:
  node1:
    text: "First node"
    invalid: [unclosed bracket
`

	err := os.WriteFile(storyPath, []byte(invalidYAML), 0o600) //nolint:gosec // test file permissions are acceptable
	if err != nil {
		t.Fatalf("Failed to create test story file: %v", err)
	}

	_, err = LoadStory(storyPath)
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestLoadStory_WithEffects(t *testing.T) {
	tmpDir := t.TempDir()
	storyPath := filepath.Join(tmpDir, "effects_story.yaml")

	storyYAML := `start: "start"
nodes:
  start:
    text: "Start"
    choices:
      - key: "heal"
        text: "Heal"
        next: "start"
        effects:
          - op: "add"
            stat: "health"
            value: 1
            clampMax: 12
`

	err := os.WriteFile(storyPath, []byte(storyYAML), 0o600) //nolint:gosec // test file permissions are acceptable
	if err != nil {
		t.Fatalf("Failed to create test story file: %v", err)
	}

	story, err := LoadStory(storyPath)
	if err != nil {
		t.Fatalf("Unexpected error loading story: %v", err)
	}

	node := story.Nodes["start"]
	if len(node.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(node.Choices))
	}

	choice := node.Choices[0]
	if len(choice.Effects) != 1 {
		t.Fatalf("Expected 1 effect, got %d", len(choice.Effects))
	}

	effect := choice.Effects[0]
	if effect.Op != "add" {
		t.Errorf("Expected op 'add', got '%s'", effect.Op)
	}
	if effect.Stat != "health" {
		t.Errorf("Expected stat 'health', got '%s'", effect.Stat)
	}
	if effect.Value != 1 {
		t.Errorf("Expected value 1, got %d", effect.Value)
	}
	if effect.ClampMax == nil || *effect.ClampMax != 12 {
		t.Errorf("Expected clampMax 12, got %v", effect.ClampMax)
	}
}
