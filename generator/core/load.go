package core

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadSpec loads and parses YAML specification from file
func LoadSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var spec Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &spec, nil
}