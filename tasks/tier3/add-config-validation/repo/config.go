package main

import (
	"encoding/json"
	"os"
)

// Config holds application configuration.
// PROBLEM: No validation, no defaults, no merge logic.
type Config struct {
	DataDir         string `json:"data_dir"`
	MaxTasks        int    `json:"max_tasks"`
	DefaultPriority string `json:"default_priority"`
	OutputFormat    string `json:"output_format"`
}

// LoadConfig loads configuration from a file.
// PROBLEM: No validation, no default values applied.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// PROBLEM: Returns config directly without validation
	// PROBLEM: No default values for missing fields
	return &cfg, nil
}
