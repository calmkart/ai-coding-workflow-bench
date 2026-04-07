package config

import (
	"os"
	"path/filepath"
	"testing"
)

// --- Task Discovery: Standard directory structure ---

func TestScenario_DiscoverTasks_MultipleTiers(t *testing.T) {
	dir := t.TempDir()

	// Create tier1 and tier2 tasks
	tasks := []struct {
		path string
		yaml string
	}{
		{
			path: filepath.Join(dir, "tier1", "fix-handler-bug"),
			yaml: `id: "tier1/fix-handler-bug"
name: "Fix handler bug"
tier: 1
type: "http-server"
language: "go"
estimated_minutes: 5
`,
		},
		{
			path: filepath.Join(dir, "tier1", "add-health-check"),
			yaml: `id: "tier1/add-health-check"
name: "Add health check"
tier: 1
type: "http-server"
language: "go"
estimated_minutes: 5
`,
		},
		{
			path: filepath.Join(dir, "tier2", "extract-storage"),
			yaml: `id: "tier2/extract-storage"
name: "Extract storage"
tier: 2
type: "http-server"
language: "go"
estimated_minutes: 10
`,
		},
	}

	for _, tt := range tasks {
		if err := os.MkdirAll(tt.path, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", tt.path, err)
		}
		if err := os.WriteFile(filepath.Join(tt.path, "task.yaml"), []byte(tt.yaml), 0644); err != nil {
			t.Fatalf("write task.yaml: %v", err)
		}
	}

	discovered, err := DiscoverTasks(dir)
	if err != nil {
		t.Fatalf("DiscoverTasks: %v", err)
	}
	if len(discovered) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(discovered))
	}

	// Verify all expected task IDs are present
	idMap := make(map[string]bool)
	for _, task := range discovered {
		idMap[task.ID] = true
	}
	for _, expectedID := range []string{"tier1/fix-handler-bug", "tier1/add-health-check", "tier2/extract-storage"} {
		if !idMap[expectedID] {
			t.Errorf("expected task %q not found in discovered tasks", expectedID)
		}
	}
}

// --- Task Discovery: Empty directory ---

func TestScenario_DiscoverTasks_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	discovered, err := DiscoverTasks(dir)
	if err != nil {
		t.Fatalf("DiscoverTasks on empty dir: %v", err)
	}
	if len(discovered) != 0 {
		t.Errorf("expected 0 tasks from empty dir, got %d", len(discovered))
	}
}

// --- Task Discovery: Directory with no task.yaml ---

func TestScenario_DiscoverTasks_NoTaskYaml(t *testing.T) {
	dir := t.TempDir()

	// Create tier structure but no task.yaml
	tierDir := filepath.Join(dir, "tier1", "some-task")
	if err := os.MkdirAll(tierDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Write a random file, not task.yaml
	if err := os.WriteFile(filepath.Join(tierDir, "readme.md"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	discovered, err := DiscoverTasks(dir)
	if err != nil {
		t.Fatalf("DiscoverTasks with no task.yaml: %v", err)
	}
	if len(discovered) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(discovered))
	}
}

// --- Filter: by tier ---

func TestScenario_FilterTasks_ByTier(t *testing.T) {
	tasks := []*TaskMeta{
		{ID: "tier1/fix-handler-bug", Tier: 1},
		{ID: "tier1/add-health-check", Tier: 1},
		{ID: "tier2/extract-storage", Tier: 2},
		{ID: "tier3/refactor-to-service", Tier: 3},
	}

	// tier1 filter
	got := FilterTasks(tasks, "tier1")
	if len(got) != 2 {
		t.Errorf("tier1 filter: expected 2, got %d", len(got))
	}
	for _, task := range got {
		if task.Tier != 1 {
			t.Errorf("tier1 filter returned task with tier %d", task.Tier)
		}
	}

	// tier2 filter
	got = FilterTasks(tasks, "tier2")
	if len(got) != 1 {
		t.Errorf("tier2 filter: expected 1, got %d", len(got))
	}

	// tier3 filter
	got = FilterTasks(tasks, "tier3")
	if len(got) != 1 {
		t.Errorf("tier3 filter: expected 1, got %d", len(got))
	}
}

// --- Filter: specific task ID ---

func TestScenario_FilterTasks_SpecificTask(t *testing.T) {
	tasks := []*TaskMeta{
		{ID: "tier1/fix-handler-bug", Tier: 1},
		{ID: "tier1/add-health-check", Tier: 1},
		{ID: "tier2/extract-storage", Tier: 2},
	}

	got := FilterTasks(tasks, "tier1/fix-handler-bug")
	if len(got) != 1 {
		t.Fatalf("specific task filter: expected 1, got %d", len(got))
	}
	if got[0].ID != "tier1/fix-handler-bug" {
		t.Errorf("expected tier1/fix-handler-bug, got %s", got[0].ID)
	}
}

// --- Filter: "all" returns everything ---

func TestScenario_FilterTasks_All(t *testing.T) {
	tasks := []*TaskMeta{
		{ID: "tier1/fix-handler-bug", Tier: 1},
		{ID: "tier2/extract-storage", Tier: 2},
		{ID: "tier3/refactor-to-service", Tier: 3},
	}

	got := FilterTasks(tasks, "all")
	if len(got) != 3 {
		t.Errorf("all filter: expected 3, got %d", len(got))
	}
}

// --- Filter: non-existent tier returns empty ---

func TestScenario_FilterTasks_NonexistentTier(t *testing.T) {
	tasks := []*TaskMeta{
		{ID: "tier1/fix-handler-bug", Tier: 1},
	}

	got := FilterTasks(tasks, "tier99")
	if len(got) != 0 {
		t.Errorf("nonexistent tier filter: expected 0, got %d", len(got))
	}
}

// --- Adversarial: Malformed task.yaml ---

func TestScenario_DiscoverTasks_MalformedTaskYaml(t *testing.T) {
	dir := t.TempDir()

	taskDir := filepath.Join(dir, "tier1", "bad-task")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Write invalid YAML
	if err := os.WriteFile(filepath.Join(taskDir, "task.yaml"), []byte("{{not valid yaml"), 0644); err != nil {
		t.Fatal(err)
	}

	// DiscoverTasks should either skip the malformed file or return an error
	// The key is it should not panic
	_, _ = DiscoverTasks(dir) // just verify no panic
}

// --- Task Dir field is populated ---

func TestScenario_DiscoverTasks_DirFieldSet(t *testing.T) {
	dir := t.TempDir()

	taskDir := filepath.Join(dir, "tier1", "test-task")
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		t.Fatal(err)
	}
	yaml := `id: "tier1/test-task"
name: "Test Task"
tier: 1
type: "http-server"
language: "go"
estimated_minutes: 5
`
	if err := os.WriteFile(filepath.Join(taskDir, "task.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	discovered, err := DiscoverTasks(dir)
	if err != nil {
		t.Fatalf("DiscoverTasks: %v", err)
	}
	if len(discovered) != 1 {
		t.Fatalf("expected 1 task, got %d", len(discovered))
	}
	if discovered[0].Dir != taskDir {
		t.Errorf("expected Dir=%q, got %q", taskDir, discovered[0].Dir)
	}
}
