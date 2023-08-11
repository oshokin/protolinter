package checker

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// MimirConfig defines the structure of the mimir file.
type MimirConfig struct {
	ProtoPaths []string `yaml:"proto_paths"`
}

func extractFilesFromMimir(file string) ([]string, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read mimir file: %w", err)
	}

	var cfg MimirConfig
	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal mimir file: %w", err)
	}

	files, err := extractFilesFromPatterns(cfg.ProtoPaths, "proto")
	if err != nil {
		return nil, fmt.Errorf("failed to extract files from \"proto_paths\" section: %w", err)
	}

	return files, nil
}
