package game

import (
	"os"
	"path/filepath"
	"testing"
)

const testStartNode = "node1"

func TestLoadStory_Valid(t *testing.T) {
	// Create a temporary YAML file
	tmpDir := t.TempDir()
	storyPath := filepath.Join(tmpDir, "test_story.yaml")

	storyYAML := `start: "` + testStartNode + `"
nodes:
  ` + testStartNode + `:
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

	if story.Start != testStartNode {
		t.Errorf("Expected start node %q, got %q", testStartNode, story.Start)
	}

	if story.Nodes == nil {
		t.Fatal("Expected nodes map to be initialized")
	}

	node1, ok := story.Nodes[testStartNode]
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

	invalidYAML := `start: "` + testStartNode + `"
nodes:
  ` + testStartNode + `:
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

func TestLoadStory_PromptChoice(t *testing.T) {
	tmpDir := t.TempDir()
	storyPath := filepath.Join(tmpDir, "prompt_story.yaml")

	storyYAML := `start: "riddle"
nodes:
  riddle:
    text: "Riddle"
    choices:
      - key: "answer"
        text: "Answer"
        prompt:
          question: "What am I?"
          placeholder: "Your answer"
          answers:
            - match: "echo"
              next: "right"
            - matches: ["shadow", "a shadow"]
              next: "wrong"
          defaultNext: "wrong"
  right:
    text: "Right"
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

	node := story.Nodes["riddle"]
	if node == nil || len(node.Choices) != 1 {
		t.Fatalf("Expected riddle node with one choice")
	}
	choice := node.Choices[0]
	if choice.Prompt == nil {
		t.Fatal("Expected prompt to be loaded")
	}
	if choice.Prompt.Question != "What am I?" {
		t.Errorf("Expected question, got %q", choice.Prompt.Question)
	}
	if choice.Prompt.Placeholder != "Your answer" {
		t.Errorf("Expected placeholder, got %q", choice.Prompt.Placeholder)
	}
	if len(choice.Prompt.Answers) != 2 {
		t.Fatalf("Expected 2 answers, got %d", len(choice.Prompt.Answers))
	}
	if choice.Prompt.Answers[0].Match != "echo" || choice.Prompt.Answers[0].Next != "right" {
		t.Errorf("Expected first answer match 'echo' -> 'right', got %+v", choice.Prompt.Answers[0])
	}
	if len(choice.Prompt.Answers[1].Matches) != 2 || choice.Prompt.Answers[1].Next != "wrong" {
		t.Errorf("Expected second answer matches -> 'wrong', got %+v", choice.Prompt.Answers[1])
	}
	if choice.Prompt.DefaultNext != "wrong" {
		t.Errorf("Expected defaultNext 'wrong', got %q", choice.Prompt.DefaultNext)
	}
}

func TestLoadStory_SceneryAndEntryAnimation(t *testing.T) {
	tmpDir := t.TempDir()
	storyPath := filepath.Join(tmpDir, "scenery_story.yaml")

	storyYAML := `start: "outside"
nodes:
  outside:
    text: "You stand before a door."
    scenery: "forest"
    choices:
      - key: "enter"
        text: "Go inside"
        next: "inside"
  inside:
    text: "You step inside."
    scenery: "house_inside"
    audio: "house_ambient"
    entry_animation: "door_open"
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

	outside := story.Nodes["outside"]
	if outside == nil {
		t.Fatal("Expected outside node to exist")
	}
	if outside.Scenery != "forest" {
		t.Errorf("Expected scenery 'forest', got %q", outside.Scenery)
	}
	if outside.EntryAnimation != "" {
		t.Errorf("Expected no entry_animation on outside, got %q", outside.EntryAnimation)
	}

	inside := story.Nodes["inside"]
	if inside == nil {
		t.Fatal("Expected inside node to exist")
	}
	if inside.Scenery != "house_inside" {
		t.Errorf("Expected scenery 'house_inside', got %q", inside.Scenery)
	}
	if inside.EntryAnimation != "door_open" {
		t.Errorf("Expected entry_animation 'door_open', got %q", inside.EntryAnimation)
	}
	if inside.Audio != "house_ambient" {
		t.Errorf("Expected audio 'house_ambient', got %q", inside.Audio)
	}
}

func TestLoadStories(t *testing.T) {
	tmpDir := t.TempDir()
	storyYAML := `start: "` + testStartNode + `"
nodes:
  ` + testStartNode + `:
    text: "First"
    ending: true
`
	err := os.WriteFile(filepath.Join(tmpDir, "one.yaml"), []byte(storyYAML), 0o600) //nolint:gosec // test dir path from t.TempDir()
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	err = os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not yaml"), 0o600) //nolint:gosec // test dir path from t.TempDir()
	if err != nil {
		t.Fatalf("Failed to create readme: %v", err)
	}

	stories, err := LoadStories(tmpDir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(stories) != 1 {
		t.Errorf("Expected 1 story, got %d", len(stories))
	}
	if stories["one"] == nil {
		t.Fatal("Expected story 'one' to exist")
	}
	if stories["one"].Start != testStartNode {
		t.Errorf("Expected start 'node1', got %q", stories["one"].Start)
	}
}

func TestLoadStories_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()
	stories, err := LoadStories(tmpDir)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(stories) != 0 {
		t.Errorf("Expected 0 stories, got %d", len(stories))
	}
}

func TestLoadStories_InvalidDir(t *testing.T) {
	_, err := LoadStories("nonexistent_directory_xyz")
	if err == nil {
		t.Error("Expected error for nonexistent directory")
	}
}
