package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Defaults.RunsPerTask != 3 {
		t.Errorf("expected runs_per_task=3, got %d", cfg.Defaults.RunsPerTask)
	}
	if cfg.Defaults.TimeoutMultiplier != 3 {
		t.Errorf("expected timeout_multiplier=3, got %d", cfg.Defaults.TimeoutMultiplier)
	}
	if _, ok := cfg.Workflows["vanilla"]; !ok {
		t.Error("expected vanilla workflow in defaults")
	}
}

func TestLoadMissingConfig(t *testing.T) {
	cfg, err := Load("/nonexistent/path/bench.yaml")
	if err != nil {
		t.Fatalf("expected no error for missing config, got: %v", err)
	}
	if cfg.Defaults.RunsPerTask != 3 {
		t.Errorf("expected default runs_per_task=3, got %d", cfg.Defaults.RunsPerTask)
	}
}

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
defaults:
  runs_per_task: 5
  timeout_multiplier: 2
`
	path := filepath.Join(dir, "bench.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Defaults.RunsPerTask != 5 {
		t.Errorf("expected runs_per_task=5, got %d", cfg.Defaults.RunsPerTask)
	}
	if cfg.Defaults.TimeoutMultiplier != 2 {
		t.Errorf("expected timeout_multiplier=2, got %d", cfg.Defaults.TimeoutMultiplier)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bench.yaml")
	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestDiscoverTasks(t *testing.T) {
	dir := t.TempDir()
	taskDir := filepath.Join(dir, "tier1", "test-task")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatal(err)
	}

	taskYAML := `
id: "tier1/test-task"
name: "Test Task"
tier: 1
type: "http-server"
language: "go"
estimated_minutes: 5
`
	if err := os.WriteFile(filepath.Join(taskDir, "task.yaml"), []byte(taskYAML), 0644); err != nil {
		t.Fatal(err)
	}

	tasks, err := DiscoverTasks(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].ID != "tier1/test-task" {
		t.Errorf("expected id 'tier1/test-task', got %q", tasks[0].ID)
	}
	if tasks[0].Tier != 1 {
		t.Errorf("expected tier 1, got %d", tasks[0].Tier)
	}
	if tasks[0].Dir != taskDir {
		t.Errorf("expected dir %q, got %q", taskDir, tasks[0].Dir)
	}
}

// TestDiscoverTasks_NonexistentDir verifies that DiscoverTasks returns an error
// when the tasks directory does not exist.
// Fix 12: Return error for nonexistent tasks directory.
func TestDiscoverTasks_NonexistentDir(t *testing.T) {
	_, err := DiscoverTasks("/nonexistent/tasks/dir")
	if err == nil {
		t.Error("expected error for nonexistent tasks directory, got nil")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected 'does not exist' in error, got: %v", err)
	}
}

// TestDefaultHomeDir_Fallback verifies DefaultHomeDir returns a valid path.
// Fix 5: Handle empty HOME gracefully.
func TestDefaultHomeDir_ReturnsValidPath(t *testing.T) {
	home := DefaultHomeDir()
	if home == "" {
		t.Error("DefaultHomeDir returned empty string")
	}
	if !filepath.IsAbs(home) {
		t.Errorf("expected absolute path, got: %s", home)
	}
}

// TestLoad_ExplicitConfigPath_IsolatesDBPath verifies that when --config is explicitly
// provided, HomeDir and DBPath are derived from the config file's directory.
func TestLoad_ExplicitConfigPath_IsolatesDBPath(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
defaults:
  runs_per_task: 5
`
	cfgPath := filepath.Join(dir, "bench.yaml")
	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HomeDir != dir {
		t.Errorf("expected HomeDir=%q, got %q", dir, cfg.HomeDir)
	}
	expectedDB := filepath.Join(dir, "results.db")
	if cfg.DBPath != expectedDB {
		t.Errorf("expected DBPath=%q, got %q", expectedDB, cfg.DBPath)
	}
}

// TestLoad_EmptyConfigPath_UsesGlobalDefaults verifies that when no --config is given,
// HomeDir and DBPath default to the global location (~/.claude/workflow-bench).
func TestLoad_EmptyConfigPath_UsesGlobalDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedHome := DefaultHomeDir()
	if cfg.HomeDir != expectedHome {
		t.Errorf("expected HomeDir=%q, got %q", expectedHome, cfg.HomeDir)
	}
	expectedDB := filepath.Join(expectedHome, "results.db")
	if cfg.DBPath != expectedDB {
		t.Errorf("expected DBPath=%q, got %q", expectedDB, cfg.DBPath)
	}
}

// TestLoad_YAMLDBPathOverride verifies that the db_path YAML field overrides the
// auto-derived DBPath.
func TestLoad_YAMLDBPathOverride(t *testing.T) {
	dir := t.TempDir()
	customDB := filepath.Join(dir, "custom", "my-results.db")
	yamlContent := `
db_path: "` + customDB + `"
defaults:
  runs_per_task: 3
`
	cfgPath := filepath.Join(dir, "bench.yaml")
	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DBPath != customDB {
		t.Errorf("expected DBPath=%q (from YAML db_path), got %q", customDB, cfg.DBPath)
	}
}

// TestLoad_YAMLWithoutDBPath_UsesAutoDerivation verifies that when the YAML file
// does not set db_path, the auto-derived path is used.
func TestLoad_YAMLWithoutDBPath_UsesAutoDerivation(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
defaults:
  runs_per_task: 7
`
	cfgPath := filepath.Join(dir, "bench.yaml")
	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectedDB := filepath.Join(dir, "results.db")
	if cfg.DBPath != expectedDB {
		t.Errorf("expected DBPath=%q, got %q", expectedDB, cfg.DBPath)
	}
}

// TestLoad_MissingExplicitConfig_StillIsolates verifies that even when the explicit
// config file does not exist (and defaults are used), the HomeDir and DBPath are
// still derived from the config path's directory.
func TestLoad_MissingExplicitConfig_StillIsolates(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "nonexistent.yaml")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.HomeDir != dir {
		t.Errorf("expected HomeDir=%q, got %q", dir, cfg.HomeDir)
	}
	expectedDB := filepath.Join(dir, "results.db")
	if cfg.DBPath != expectedDB {
		t.Errorf("expected DBPath=%q, got %q", expectedDB, cfg.DBPath)
	}
}

func TestFilterTasks(t *testing.T) {
	tasks := []*TaskMeta{
		{ID: "tier1/fix-handler-bug", Tier: 1},
		{ID: "tier1/add-health-check", Tier: 1},
		{ID: "tier2/extract-storage", Tier: 2},
	}

	tests := []struct {
		selector string
		wantLen  int
	}{
		{"all", 3},
		{"tier1", 2},
		{"tier2", 1},
		{"tier1/fix-handler-bug", 1},
		{"tier3", 0},
	}

	for _, tt := range tests {
		got := FilterTasks(tasks, tt.selector)
		if len(got) != tt.wantLen {
			t.Errorf("FilterTasks(%q): expected %d tasks, got %d", tt.selector, tt.wantLen, len(got))
		}
	}
}

// TestLoadJudgeConfig verifies that the judge section is parsed from YAML.
func TestLoadJudgeConfig(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
judge:
  enabled: true
  model: claude-sonnet-4-20250514
defaults:
  runs_per_task: 3
`
	path := filepath.Join(dir, "bench.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Judge.Enabled {
		t.Error("expected judge.enabled=true")
	}
	if cfg.Judge.Model != "claude-sonnet-4-20250514" {
		t.Errorf("expected judge.model=claude-sonnet-4-20250514, got %q", cfg.Judge.Model)
	}
}

// TestDiscoverTasks_VTsParsed verifies that verification_targets are parsed from task.yaml.
func TestDiscoverTasks_VTsParsed(t *testing.T) {
	dir := t.TempDir()
	taskDir := filepath.Join(dir, "tier1", "test-vt")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatal(err)
	}

	taskYAML := `
id: "tier1/test-vt"
tier: 1
type: "concurrency"
estimated_minutes: 5
verification_targets:
  - id: VT-LEAK-01
    category: concurrency
    name: "goroutine leak"
    severity: critical
    detection: "e2e test case"
  - id: VT-BUILD-01
    category: build
    name: "build check"
    severity: medium
    detection: "go build"
`
	if err := os.WriteFile(filepath.Join(taskDir, "task.yaml"), []byte(taskYAML), 0644); err != nil {
		t.Fatal(err)
	}

	tasks, err := DiscoverTasks(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if len(tasks[0].VerificationTargets) != 2 {
		t.Fatalf("expected 2 VTs, got %d", len(tasks[0].VerificationTargets))
	}
	vt0 := tasks[0].VerificationTargets[0]
	if vt0.ID != "VT-LEAK-01" {
		t.Errorf("expected VT id 'VT-LEAK-01', got %q", vt0.ID)
	}
	if vt0.Severity != "critical" {
		t.Errorf("expected severity 'critical', got %q", vt0.Severity)
	}
	if vt0.Detection != "e2e test case" {
		t.Errorf("expected detection 'e2e test case', got %q", vt0.Detection)
	}
}

// TestLoadJudgeConfig_DefaultsToDisabled verifies judge is disabled by default.
func TestLoadJudgeConfig_DefaultsToDisabled(t *testing.T) {
	dir := t.TempDir()
	yamlContent := `
defaults:
  runs_per_task: 3
`
	path := filepath.Join(dir, "bench.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Judge.Enabled {
		t.Error("expected judge.enabled=false by default")
	}
	if cfg.Judge.Model != "" {
		t.Errorf("expected empty judge.model by default, got %q", cfg.Judge.Model)
	}
}
