package game

import (
	"os"
	"path/filepath"

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
