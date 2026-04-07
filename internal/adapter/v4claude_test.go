package adapter

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewV4Claude_DefaultAgentsDir(t *testing.T) {
	a, err := NewV4Claude(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v4 := a.(*V4ClaudeAdapter)
	if v4.AgentsDir == "" {
		t.Error("expected non-empty AgentsDir")
	}
	// Should end with .claude/agents
	if filepath.Base(filepath.Dir(v4.AgentsDir)) != ".claude" || filepath.Base(v4.AgentsDir) != "agents" {
		t.Errorf("expected AgentsDir to end with .claude/agents, got %s", v4.AgentsDir)
	}
}

func TestNewV4Claude_ConfigOverride(t *testing.T) {
	cfg := map[string]any{
		"agents_dir": "/custom/agents",
	}
	a, err := NewV4Claude(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v4 := a.(*V4ClaudeAdapter)
	if v4.AgentsDir != "/custom/agents" {
		t.Errorf("expected /custom/agents, got %s", v4.AgentsDir)
	}
}

func TestNewV4Claude_TildeExpansion(t *testing.T) {
	cfg := map[string]any{
		"agents_dir": "~/.claude/agents",
	}
	a, err := NewV4Claude(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v4 := a.(*V4ClaudeAdapter)
	// Should not start with ~
	if len(v4.AgentsDir) > 0 && v4.AgentsDir[0] == '~' {
		t.Errorf("expected tilde to be expanded, got %s", v4.AgentsDir)
	}
}

func TestV4ClaudeAdapter_Name(t *testing.T) {
	a := &V4ClaudeAdapter{AgentsDir: "/tmp"}
	if a.Name() != "v4-claude" {
		t.Errorf("expected name 'v4-claude', got %q", a.Name())
	}
}

func TestV4ClaudeAdapter_Setup(t *testing.T) {
	// Create a fake agents source directory with .md files and a reference/ subdir.
	agentsDir := t.TempDir()

	// Create agent .md files.
	for _, name := range []string{"manager.md", "coding.md", "testing.md"} {
		content := "# " + name + "\nSome agent instructions."
		if err := os.WriteFile(filepath.Join(agentsDir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create a non-.md file that should NOT be copied.
	if err := os.WriteFile(filepath.Join(agentsDir, "notes.txt"), []byte("skip me"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create reference/ subdirectory with a file.
	refDir := filepath.Join(agentsDir, "reference")
	if err := os.MkdirAll(refDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(refDir, "security.md"), []byte("# Security rules"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a non-reference directory that should NOT be copied.
	otherDir := filepath.Join(agentsDir, "other-dir")
	if err := os.MkdirAll(otherDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(otherDir, "file.md"), []byte("skip me too"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a worktree directory.
	worktreeDir := t.TempDir()

	a := &V4ClaudeAdapter{AgentsDir: agentsDir}
	if err := a.Setup(context.Background(), worktreeDir); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Verify agent .md files were copied.
	agentsTarget := filepath.Join(worktreeDir, ".claude", "agents")
	for _, name := range []string{"manager.md", "coding.md", "testing.md"} {
		path := filepath.Join(agentsTarget, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected agent file %s to be copied", name)
		}
	}

	// Verify non-.md file was NOT copied.
	if _, err := os.Stat(filepath.Join(agentsTarget, "notes.txt")); !os.IsNotExist(err) {
		t.Error("expected notes.txt to NOT be copied")
	}

	// Verify reference/ was copied.
	refTarget := filepath.Join(agentsTarget, "reference", "security.md")
	if _, err := os.Stat(refTarget); os.IsNotExist(err) {
		t.Error("expected reference/security.md to be copied")
	}

	// Verify other-dir was NOT copied.
	otherTarget := filepath.Join(agentsTarget, "other-dir")
	if _, err := os.Stat(otherTarget); !os.IsNotExist(err) {
		t.Error("expected other-dir to NOT be copied")
	}

	// Verify .planning/manager/ was created.
	planningDir := filepath.Join(worktreeDir, ".planning", "manager")
	info, err := os.Stat(planningDir)
	if os.IsNotExist(err) {
		t.Error("expected .planning/manager/ directory to be created")
	} else if !info.IsDir() {
		t.Error("expected .planning/manager/ to be a directory")
	}
}

func TestV4ClaudeAdapter_Setup_MissingAgentsDir(t *testing.T) {
	worktreeDir := t.TempDir()
	a := &V4ClaudeAdapter{AgentsDir: "/nonexistent/agents/dir"}
	err := a.Setup(context.Background(), worktreeDir)
	if err == nil {
		t.Error("expected error for nonexistent agents dir")
	}
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir")
	}

	tests := []struct {
		input string
		want  string
	}{
		{"~/.claude/agents", filepath.Join(home, ".claude", "agents")},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~", "~"}, // Only "~/" prefix is expanded, not bare "~"
	}

	for _, tt := range tests {
		got := expandHome(tt.input)
		if got != tt.want {
			t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestCopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	dst := filepath.Join(dir, "dest.txt")

	content := "hello world"
	if err := os.WriteFile(src, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(data) != content {
		t.Errorf("expected %q, got %q", content, string(data))
	}
}

func TestCopyDir(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "copy")

	// Create nested structure.
	subDir := filepath.Join(srcDir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "a.txt"), []byte("a"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "b.txt"), []byte("b"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := copyDir(srcDir, dstDir); err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	// Verify files.
	data, err := os.ReadFile(filepath.Join(dstDir, "a.txt"))
	if err != nil {
		t.Fatalf("read a.txt: %v", err)
	}
	if string(data) != "a" {
		t.Errorf("a.txt: expected 'a', got %q", string(data))
	}

	data, err = os.ReadFile(filepath.Join(dstDir, "sub", "b.txt"))
	if err != nil {
		t.Fatalf("read sub/b.txt: %v", err)
	}
	if string(data) != "b" {
		t.Errorf("sub/b.txt: expected 'b', got %q", string(data))
	}
}

func TestRegistryContainsV4Claude(t *testing.T) {
	if _, ok := Registry["v4-claude"]; !ok {
		t.Error("expected v4-claude in adapter Registry")
	}
}

func TestGetV4Claude(t *testing.T) {
	a, err := Get("v4-claude", map[string]any{"agents_dir": "/tmp/test-agents"})
	if err != nil {
		t.Fatalf("Get v4-claude failed: %v", err)
	}
	if a.Name() != "v4-claude" {
		t.Errorf("expected name 'v4-claude', got %q", a.Name())
	}
}
