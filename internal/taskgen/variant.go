// Package taskgen implements dynamic task variant generation.
// It creates new task variants by copying a source task and applying
// simple string replacements to create semantically equivalent but
// syntactically different versions.
//
// Usage: workflow-bench generate-variant --source tier1/fix-handler-bug --seed 42
package taskgen

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// VariantConfig holds parameters for generating a task variant.
type VariantConfig struct {
	SourceTask string // Source task directory path
	Seed       int64  // Random seed (0 = random)
	OutputDir  string // Output directory for the variant
}

// replacementSets defines the string replacement groups.
// Each key maps to a list of possible replacements.
// When generating a variant, one replacement is chosen per group.
var replacementSets = map[string][]string{
	"Todo":  {"Item", "Task", "Note", "Entry", "Record"},
	"todo":  {"item", "task", "note", "entry", "record"},
	"todos": {"items", "tasks", "notes", "entries", "records"},
	"Todos": {"Items", "Tasks", "Notes", "Entries", "Records"},
}

// GenerateVariant creates a variant of the source task by copying all files
// and applying string replacements to .go, .md, and .yaml files.
//
// @implements P21 (Dynamic task variant generation)
func GenerateVariant(cfg VariantConfig) error {
	if cfg.SourceTask == "" {
		return fmt.Errorf("source task path cannot be empty")
	}
	if cfg.OutputDir == "" {
		return fmt.Errorf("output directory cannot be empty")
	}

	// Validate source exists.
	srcAbs, err := filepath.Abs(cfg.SourceTask)
	if err != nil {
		return fmt.Errorf("resolve source path: %w", err)
	}
	if _, err := os.Stat(srcAbs); err != nil {
		return fmt.Errorf("source task not found: %s", srcAbs)
	}

	// Initialize random generator.
	rng := rand.New(rand.NewSource(cfg.Seed))

	// Pick a replacement index (all groups use the same index for consistency).
	replacementIdx := rng.Intn(len(replacementSets["Todo"]))

	// Build the replacement map for this variant.
	replMap := buildReplacementMap(replacementIdx)

	// Create output directory.
	outAbs, err := filepath.Abs(cfg.OutputDir)
	if err != nil {
		return fmt.Errorf("resolve output path: %w", err)
	}

	// Copy the entire source tree.
	if err := copyDir(srcAbs, outAbs); err != nil {
		return fmt.Errorf("copy source: %w", err)
	}

	// Apply replacements to text files.
	if err := applyReplacements(outAbs, replMap); err != nil {
		return fmt.Errorf("apply replacements: %w", err)
	}

	// Update task.yaml ID to include variant suffix.
	if err := updateTaskID(outAbs, replacementIdx); err != nil {
		return fmt.Errorf("update task ID: %w", err)
	}

	// Initialize git repo and create initial commit.
	repoDir := filepath.Join(outAbs, "repo")
	if _, err := os.Stat(repoDir); err == nil {
		if err := initGitRepo(repoDir); err != nil {
			return fmt.Errorf("init git in repo/: %w", err)
		}
	}

	return nil
}

// buildReplacementMap creates a from->to mapping using the given replacement index.
func buildReplacementMap(idx int) map[string]string {
	m := make(map[string]string)
	for key, alternatives := range replacementSets {
		if idx < len(alternatives) {
			m[key] = alternatives[idx]
		}
	}
	return m
}

// copyDir recursively copies a directory tree.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Compute relative path.
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(dst, relPath)

		// Skip .git directories in source.
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		// Copy file.
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		return os.WriteFile(destPath, data, info.Mode())
	})
}

// applyReplacements walks the output directory and applies string replacements
// to files with recognized text extensions (.go, .md, .yaml, .src).
func applyReplacements(dir string, replMap map[string]string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Only process text files.
		ext := filepath.Ext(path)
		if !isTextFile(ext, info.Name()) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		content := string(data)
		modified := false

		// Apply replacements in a deterministic order (longer keys first to avoid partial matches).
		// Order: "Todos" > "todos" > "Todo" > "todo"
		orderedKeys := []string{"Todos", "todos", "Todo", "todo"}
		for _, key := range orderedKeys {
			if replacement, ok := replMap[key]; ok {
				if strings.Contains(content, key) {
					content = strings.ReplaceAll(content, key, replacement)
					modified = true
				}
			}
		}

		if modified {
			if err := os.WriteFile(path, []byte(content), info.Mode()); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
		}

		return nil
	})
}

// isTextFile checks if a file should have replacements applied.
func isTextFile(ext, name string) bool {
	switch ext {
	case ".go", ".md", ".yaml", ".yml", ".src", ".txt", ".json":
		return true
	}
	// Also handle files like "e2e_test.go.src".
	if strings.HasSuffix(name, ".go.src") {
		return true
	}
	return false
}

// updateTaskID reads task.yaml and appends a variant suffix to the ID.
func updateTaskID(dir string, variantIdx int) error {
	taskPath := filepath.Join(dir, "task.yaml")
	data, err := os.ReadFile(taskPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No task.yaml, nothing to update.
		}
		return err
	}

	content := string(data)
	variantSuffix := fmt.Sprintf("-variant-%d", variantIdx)

	// Simple string replacement: append suffix to the id field.
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "id:") {
			// Extract current ID value.
			idVal := strings.TrimSpace(strings.TrimPrefix(trimmed, "id:"))
			idVal = strings.Trim(idVal, `"'`)
			newID := idVal + variantSuffix
			lines[i] = fmt.Sprintf("id: %q", newID)
			break
		}
	}

	return os.WriteFile(taskPath, []byte(strings.Join(lines, "\n")), 0644)
}

// initGitRepo initializes a fresh git repo with an initial commit.
func initGitRepo(dir string) error {
	// Remove existing .git if any.
	gitDir := filepath.Join(dir, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		return err
	}

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "bench@example.com"},
		{"git", "config", "user.name", "workflow-bench"},
		{"git", "add", "-A"},
		{"git", "commit", "-m", "variant initial commit"},
	}

	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("%v: %s", args, string(out))
		}
	}
	return nil
}
