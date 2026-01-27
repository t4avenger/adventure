package game

import (
	"os"

	"gopkg.in/yaml.v3"
)

func LoadStory(path string) (*Story, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Story
	if err := yaml.Unmarshal(b, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
