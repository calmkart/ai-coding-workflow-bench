package taskgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildReplacementMap(t *testing.T) {
	m := buildReplacementMap(0)
	if m["Todo"] != "Item" {
		t.Errorf("expected Todo -> Item, got %q", m["Todo"])
	}
	if m["todo"] != "item" {
		t.Errorf("expected todo -> item, got %q", m["todo"])
	}
	if m["todos"] != "items" {
		t.Errorf("expected todos -> items, got %q", m["todos"])
	}

	m2 := buildReplacementMap(2)
	if m2["Todo"] != "Note" {
		t.Errorf("expected Todo -> Note for idx=2, got %q", m2["Todo"])
	}
}

func TestIsTextFile(t *testing.T) {
	tests := []struct {
		ext  string
		name string
		want bool
	}{
		{".go", "main.go", true},
		{".md", "plan.md", true},
		{".yaml", "task.yaml", true},
		{".src", "e2e_test.go.src", true},
		{".bin", "data.bin", false},
		{".png", "image.png", false},
	}
	for _, tt := range tests {
		got := isTextFile(tt.ext, tt.name)
		if got != tt.want {
			t.Errorf("isTextFile(%q, %q) = %v, want %v", tt.ext, tt.name, got, tt.want)
		}
	}
}

func TestCopyDir(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "copy")

	// Create source structure.
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}
	// Create .git dir that should be skipped.
	if err := os.MkdirAll(filepath.Join(src, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, ".git", "config"), []byte("git"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyDir(src, dst); err != nil {
		t.Fatal(err)
	}

	// Verify files were copied.
	data, err := os.ReadFile(filepath.Join(dst, "a.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Errorf("a.txt content = %q, want %q", string(data), "hello")
	}

	data, err = os.ReadFile(filepath.Join(dst, "sub", "b.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "world" {
		t.Errorf("sub/b.txt content = %q, want %q", string(data), "world")
	}

	// Verify .git was skipped.
	if _, err := os.Stat(filepath.Join(dst, ".git")); !os.IsNotExist(err) {
		t.Error(".git directory should not be copied")
	}
}

func TestApplyReplacements(t *testing.T) {
	dir := t.TempDir()

	// Create a .go file with "Todo" references.
	goContent := `package main

type Todo struct {
	ID   int
	Name string
}

func handleTodos() []Todo {
	var todos []Todo
	return todos
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(goContent), 0644); err != nil {
		t.Fatal(err)
	}

	replMap := buildReplacementMap(0) // Todo -> Item, todos -> items

	if err := applyReplacements(dir, replMap); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if strings.Contains(content, "Todo") {
		t.Error("expected all 'Todo' to be replaced")
	}
	if strings.Contains(content, "todos") {
		t.Error("expected all 'todos' to be replaced")
	}
	if !strings.Contains(content, "Item") {
		t.Error("expected 'Item' to appear in replacement")
	}
	if !strings.Contains(content, "items") {
		t.Error("expected 'items' to appear in replacement")
	}
}

func TestUpdateTaskID(t *testing.T) {
	dir := t.TempDir()
	taskYAML := `id: "tier1/fix-handler-bug"
tier: 1
type: "http-server"
`
	if err := os.WriteFile(filepath.Join(dir, "task.yaml"), []byte(taskYAML), 0644); err != nil {
		t.Fatal(err)
	}

	if err := updateTaskID(dir, 2); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "task.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "variant-2") {
		t.Errorf("expected task ID to contain 'variant-2', got:\n%s", content)
	}
}

func TestUpdateTaskID_NoFile(t *testing.T) {
	dir := t.TempDir()
	// Should not error when task.yaml doesn't exist.
	if err := updateTaskID(dir, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateVariant_EmptySource(t *testing.T) {
	err := GenerateVariant(VariantConfig{SourceTask: "", OutputDir: "/tmp/out"})
	if err == nil {
		t.Fatal("expected error for empty source")
	}
}

func TestGenerateVariant_EmptyOutput(t *testing.T) {
	err := GenerateVariant(VariantConfig{SourceTask: "/tmp/src", OutputDir: ""})
	if err == nil {
		t.Fatal("expected error for empty output")
	}
}

func TestGenerateVariant_SourceNotFound(t *testing.T) {
	err := GenerateVariant(VariantConfig{
		SourceTask: "/nonexistent/path/to/task",
		OutputDir:  "/tmp/out",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent source")
	}
}

func TestGenerateVariant_FullFlow(t *testing.T) {
	// Create a source task directory.
	src := t.TempDir()
	goContent := `package main

type Todo struct { Name string }

var todos []Todo

func handleTodos() {}
`
	if err := os.WriteFile(filepath.Join(src, "main.go"), []byte(goContent), 0644); err != nil {
		t.Fatal(err)
	}
	taskYAML := `id: "test-task"
tier: 1
type: "http-server"
`
	if err := os.WriteFile(filepath.Join(src, "task.yaml"), []byte(taskYAML), 0644); err != nil {
		t.Fatal(err)
	}
	planMD := `# Task: Fix todos handler
Handle todos properly.
`
	if err := os.WriteFile(filepath.Join(src, "plan.md"), []byte(planMD), 0644); err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(t.TempDir(), "variant")

	err := GenerateVariant(VariantConfig{
		SourceTask: src,
		Seed:       42,
		OutputDir:  out,
	})
	if err != nil {
		t.Fatalf("GenerateVariant failed: %v", err)
	}

	// Check that main.go was transformed.
	data, err := os.ReadFile(filepath.Join(out, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	// With seed 42, some replacement was applied.
	if strings.Contains(content, "Todo") && strings.Contains(content, "todos") {
		t.Error("expected at least some replacements to be applied")
	}

	// Check task.yaml was updated with variant suffix.
	taskData, err := os.ReadFile(filepath.Join(out, "task.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(taskData), "variant-") {
		t.Error("expected task.yaml to contain variant suffix")
	}
}
