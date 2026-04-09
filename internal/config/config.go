// Package config loads and merges bench.yaml configuration.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the full bench configuration.
type Config struct {
	Workflows map[string]WorkflowConfig `yaml:"workflows"`
	Defaults  Defaults                  `yaml:"defaults"`
	Judge     JudgeConfig               `yaml:"judge"`

	// DBPath can be set via YAML (db_path field) or auto-derived from config location.
	// When set in YAML, it overrides the auto-derived path.
	DBPath string `yaml:"db_path,omitempty"`

	// Runtime overrides (not from YAML)
	TasksDir string `yaml:"-"`
	HomeDir  string `yaml:"-"`
}

// JudgeConfig controls the LLM-as-Judge code quality evaluation.
// When Enabled is true, each run's code changes are scored by an LLM
// using the rubric dimensions (correctness, readability, etc.).
// Requires ANTHROPIC_API_KEY environment variable.
//
// InputPricePerMTok and OutputPricePerMTok configure the cost model
// for estimating USD cost from token counts. Defaults are $3/M input
// and $15/M output (Anthropic Claude Sonnet 4 pricing).
type JudgeConfig struct {
	Enabled            bool    `yaml:"enabled"`
	Model              string  `yaml:"model"`
	InputPricePerMTok  float64 `yaml:"input_price_per_mtok"`
	OutputPricePerMTok float64 `yaml:"output_price_per_mtok"`
}

// WorkflowConfig defines a workflow adapter and its settings.
type WorkflowConfig struct {
	Adapter       string   `yaml:"adapter"`
	AgentsDir     string   `yaml:"agents_dir,omitempty"`
	EntryCommand  string   `yaml:"entry_command,omitempty"`
	SetupCommands []string `yaml:"setup_commands,omitempty"`
}

// ToMap converts WorkflowConfig fields to a map for adapter constructors.
func (wc WorkflowConfig) ToMap() map[string]any {
	m := map[string]any{
		"adapter": wc.Adapter,
	}
	if wc.AgentsDir != "" {
		m["agents_dir"] = wc.AgentsDir
	}
	if wc.EntryCommand != "" {
		m["entry_command"] = wc.EntryCommand
	}
	if len(wc.SetupCommands) > 0 {
		m["setup_commands"] = wc.SetupCommands
	}
	return m
}

// Defaults holds default run parameters.
type Defaults struct {
	RunsPerTask       int `yaml:"runs_per_task"`
	TimeoutMultiplier int `yaml:"timeout_multiplier"`
}

// VerificationTarget represents a single verification target from task.yaml.
type VerificationTarget struct {
	ID        string `yaml:"id"`
	Category  string `yaml:"category"`
	Name      string `yaml:"name"`
	Severity  string `yaml:"severity"`  // "critical", "high", "medium", "low"
	Detection string `yaml:"detection"` // "go build", "e2e test case", "errcheck", etc.
}

// TaskMeta represents a task.yaml file.
type TaskMeta struct {
	ID                  string               `yaml:"id"`
	Tier                int                  `yaml:"tier"`
	Type                string               `yaml:"type"`
	Language            string               `yaml:"language"` // "go", "python", "typescript"; defaults to "go" if empty
	EstimatedMinutes    int                  `yaml:"estimated_minutes"`
	VerificationTargets []VerificationTarget `yaml:"verification_targets"`

	// Runtime fields (not from YAML)
	Dir string `yaml:"-"` // absolute path to task directory
}

// DefaultConfig returns the built-in default configuration.
func DefaultConfig() *Config {
	return &Config{
		Workflows: map[string]WorkflowConfig{
			"vanilla": {Adapter: "vanilla"},
		},
		Defaults: Defaults{
			RunsPerTask:       3,
			TimeoutMultiplier: 3,
		},
		Judge: JudgeConfig{
			InputPricePerMTok:  3.0,
			OutputPricePerMTok: 15.0,
		},
	}
}

// DefaultHomeDir returns ~/.claude/workflow-bench.
// Fix 5: Handle os.UserHomeDir() error — fall back to /tmp if $HOME is unset.
func DefaultHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = os.TempDir()
	}
	return filepath.Join(home, ".claude", "workflow-bench")
}

// Load reads the configuration from the given path, merging with defaults.
// If configPath is empty, it uses the default global location (~/.claude/workflow-bench).
// If configPath is explicitly set, HomeDir and DBPath are derived from the config
// file's directory, providing full isolation per config location.
//
// The db_path YAML field, if set, overrides the auto-derived DBPath.
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	if configPath != "" {
		// User explicitly specified a config file -> isolate to its directory.
		absPath, err := filepath.Abs(configPath)
		if err != nil {
			return nil, fmt.Errorf("resolve config path: %w", err)
		}
		cfg.HomeDir = filepath.Dir(absPath)
		cfg.DBPath = filepath.Join(cfg.HomeDir, "results.db")
	} else {
		// Default: global location.
		cfg.HomeDir = DefaultHomeDir()
		cfg.DBPath = filepath.Join(cfg.HomeDir, "results.db")
		configPath = filepath.Join(cfg.HomeDir, "bench.yaml")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file, use defaults
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	// Remember the auto-derived DBPath before YAML parsing may overwrite it.
	autoDB := cfg.DBPath

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", configPath, err)
	}

	// If YAML did not set db_path, restore the auto-derived value
	// (yaml.Unmarshal with omitempty leaves the zero value "" when the field is absent).
	if cfg.DBPath == "" {
		cfg.DBPath = autoDB
	}

	// Default pricing if not set in YAML.
	if cfg.Judge.InputPricePerMTok == 0 {
		cfg.Judge.InputPricePerMTok = 3.0
	}
	if cfg.Judge.OutputPricePerMTok == 0 {
		cfg.Judge.OutputPricePerMTok = 15.0
	}

	return cfg, nil
}

// DiscoverTasks scans the tasks directory for task.yaml files.
// It matches the pattern tasks/tier*/*/task.yaml.
//
// @implements REQ-TASK-DISCOVER (task discovery by scanning tier*/*/task.yaml)
func DiscoverTasks(tasksDir string) ([]*TaskMeta, error) {
	// Fix 12: Return error if tasks directory does not exist.
	if _, err := os.Stat(tasksDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("tasks directory does not exist: %s (hint: use --tasks-dir to specify the path, or run from the project root)", tasksDir)
	}

	pattern := filepath.Join(tasksDir, "tier*", "*", "task.yaml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob tasks: %w", err)
	}

	var tasks []*TaskMeta
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read task %s: %w", path, err)
		}

		var task TaskMeta
		if err := yaml.Unmarshal(data, &task); err != nil {
			return nil, fmt.Errorf("parse task %s: %w", path, err)
		}
		task.Dir = filepath.Dir(path)
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

// FilterTasks filters tasks by a selector string.
// Supported selectors:
//   - "all" -> all tasks
//   - "tier1" -> all tier 1 tasks
//   - "tier1/fix-handler-bug" -> specific task
func FilterTasks(tasks []*TaskMeta, selector string) []*TaskMeta {
	if selector == "all" {
		return tasks
	}

	var filtered []*TaskMeta
	for _, t := range tasks {
		if t.ID == selector {
			return []*TaskMeta{t}
		}
		tierPrefix := fmt.Sprintf("tier%d", t.Tier)
		if strings.EqualFold(selector, tierPrefix) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// LoadPlan reads the plan.md from a task directory.
func LoadPlan(taskDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(taskDir, "plan.md"))
	if err != nil {
		return "", fmt.Errorf("read plan: %w", err)
	}
	return string(data), nil
}

