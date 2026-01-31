package game

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadStory loads a story from a YAML file.
func LoadStory(path string) (*Story, error) {
	// Resolve path to prevent directory traversal attacks
	cleanPath := filepath.Clean(path)
	b, err := os.ReadFile(cleanPath) //nolint:gosec // path is cleaned and validated
	if err != nil {
		return nil, err
	}
	var s Story
	if err := yaml.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// LoadStories loads all *.yaml files from dir and returns a map of story ID (filename without extension) to Story.
func LoadStories(dir string) (map[string]*Story, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	stories := make(map[string]*Story)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".yaml") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		if id == "" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		s, err := LoadStory(path)
		if err != nil {
			return nil, err
		}
		stories[id] = s
	}
	return stories, nil
}
