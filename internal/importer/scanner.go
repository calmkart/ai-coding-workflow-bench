// Package importer implements semi-automatic task creation from git history.
// It scans a git repository's commit range, extracts changed files, and creates
// a workflow-bench task directory structure with auto-detected type and tier.
//
// Usage: workflow-bench import --repo /path/to/project --commit abc123..def456
package importer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ImportConfig holds parameters for importing a task from git history.
type ImportConfig struct {
	RepoPath    string // Path to the git repository
	CommitRange string // Git commit range, e.g. "abc123..def456"
	Tier        int    // Override tier (0 = auto-detect)
	Type        string // Override type ("" = auto-detect)
	OutputDir   string // Output directory for the imported task (default: tasks/imported/<name>/)
}

// ImportResult holds the result of a successful import.
type ImportResult struct {
	TaskID   string   // Generated task ID
	TaskDir  string   // Absolute path to the created task directory
	DiffSize int      // Number of diff lines
	Files    []string // List of changed files
	Tier     int      // Final tier (auto-detected or overridden)
	Type     string   // Final type (auto-detected or overridden)
}

// Import creates a workflow-bench task from a git commit range.
// It extracts the diff, detects the task type and tier, and creates
// the task directory structure with a task.yaml and plan.md template.
//
// @implements P20 (Git History Importer)
func Import(cfg ImportConfig) (*ImportResult, error) {
	if cfg.RepoPath == "" {
		return nil, fmt.Errorf("repo path cannot be empty")
	}
	if cfg.CommitRange == "" {
		return nil, fmt.Errorf("commit range cannot be empty")
	}

	// Validate the repo path exists and is a git repo.
	if _, err := os.Stat(filepath.Join(cfg.RepoPath, ".git")); err != nil {
		return nil, fmt.Errorf("not a git repository: %s", cfg.RepoPath)
	}

	// Parse commit range to get the start commit.
	parts := strings.SplitN(cfg.CommitRange, "..", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("invalid commit range %q: expected format abc123..def456", cfg.CommitRange)
	}
	startCommit := parts[0]

	// Get diff content.
	diffContent, err := gitDiff(cfg.RepoPath, cfg.CommitRange)
	if err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}
	if diffContent == "" {
		return nil, fmt.Errorf("empty diff for range %s", cfg.CommitRange)
	}

	// Get changed files.
	files, err := gitDiffFiles(cfg.RepoPath, cfg.CommitRange)
	if err != nil {
		return nil, fmt.Errorf("git diff files: %w", err)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files changed in range %s", cfg.CommitRange)
	}

	// Count diff lines.
	diffLines := countDiffLines(diffContent)

	// Auto-detect type if not overridden.
	taskType := cfg.Type
	if taskType == "" {
		taskType = detectTaskType(files, diffContent)
	}

	// Auto-detect tier if not overridden.
	tier := cfg.Tier
	if tier == 0 {
		tier = estimateTier(diffLines, len(files))
	}

	// Generate task name from the commit range (sanitized).
	taskName := sanitizeTaskName(cfg.CommitRange)
	taskID := fmt.Sprintf("imported/%s", taskName)

	// Determine output directory.
	outputDir := cfg.OutputDir
	if outputDir == "" {
		outputDir = filepath.Join("tasks", "imported", taskName)
	}

	// Create directory structure.
	absOutput, err := filepath.Abs(outputDir)
	if err != nil {
		return nil, fmt.Errorf("resolve output path: %w", err)
	}

	if err := os.MkdirAll(absOutput, 0755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	// Create repo/ via git worktree or clone at start commit.
	repoDir := filepath.Join(absOutput, "repo")
	if err := createRepoSnapshot(cfg.RepoPath, startCommit, repoDir); err != nil {
		return nil, fmt.Errorf("create repo snapshot: %w", err)
	}

	// Create verify/ directory (empty, user needs to add E2E tests).
	verifyDir := filepath.Join(absOutput, "verify")
	if err := os.MkdirAll(verifyDir, 0755); err != nil {
		return nil, fmt.Errorf("create verify directory: %w", err)
	}

	// Estimate minutes based on tier.
	estimatedMinutes := tierToMinutes(tier)

	// Write task.yaml.
	taskYAML := taskYAMLContent{
		ID:               taskID,
		Tier:             tier,
		Type:             taskType,
		Language:         detectLanguage(files),
		EstimatedMinutes: estimatedMinutes,
	}
	if err := writeTaskYAML(filepath.Join(absOutput, "task.yaml"), taskYAML); err != nil {
		return nil, fmt.Errorf("write task.yaml: %w", err)
	}

	// Write plan.md template (user must edit).
	if err := writePlanTemplate(filepath.Join(absOutput, "plan.md")); err != nil {
		return nil, fmt.Errorf("write plan.md: %w", err)
	}

	return &ImportResult{
		TaskID:   taskID,
		TaskDir:  absOutput,
		DiffSize: diffLines,
		Files:    files,
		Tier:     tier,
		Type:     taskType,
	}, nil
}

// detectLanguage infers the primary programming language from the list of changed files.
// It counts file extensions and returns the language with the most files.
// Defaults to "go" when no clear winner or no recognized extensions.
func detectLanguage(files []string) string {
	goCount, pyCount, tsCount := 0, 0, 0
	for _, f := range files {
		switch {
		case strings.HasSuffix(f, ".go"):
			goCount++
		case strings.HasSuffix(f, ".py"):
			pyCount++
		case strings.HasSuffix(f, ".ts") || strings.HasSuffix(f, ".tsx"):
			tsCount++
		}
	}
	if pyCount > goCount && pyCount > tsCount {
		return "python"
	}
	if tsCount > goCount && tsCount > pyCount {
		return "typescript"
	}
	return "go"
}

// detectTaskType examines the changed files and diff content to classify the task type.
//
// @implements P20 (auto type detection)
func detectTaskType(files []string, diffContent string) string {
	hasHTTP := strings.Contains(diffContent, "http.") || strings.Contains(diffContent, "gin.")
	hasMain := containsFile(files, "main.go") || containsAnyPrefix(files, "cmd/")
	hasTest := containsAnySuffix(files, "_test.go")
	hasReconciler := strings.Contains(diffContent, "Reconcile") || strings.Contains(diffContent, "controller-runtime")

	if hasReconciler {
		return "reconciler"
	}
	if hasHTTP && hasMain {
		return "http-server"
	}
	if hasMain {
		return "cli"
	}
	if hasTest {
		return "library"
	}
	return "library"
}

// estimateTier estimates task complexity tier based on diff size and file count.
//
// @implements P20 (auto tier estimation)
func estimateTier(diffLines int, fileCount int) int {
	if diffLines < 50 && fileCount <= 2 {
		return 1
	}
	if diffLines < 200 && fileCount <= 5 {
		return 2
	}
	if diffLines < 500 {
		return 3
	}
	return 4
}

// gitDiff runs git diff on the given commit range and returns the output.
func gitDiff(repoPath, commitRange string) (string, error) {
	cmd := exec.Command("git", "diff", commitRange)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git diff failed: %s", string(exitErr.Stderr))
		}
		return "", err
	}
	return string(out), nil
}

// gitDiffFiles returns the list of files changed in the given commit range.
func gitDiffFiles(repoPath, commitRange string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", commitRange)
	cmd.Dir = repoPath
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git diff --name-only failed: %s", string(exitErr.Stderr))
		}
		return nil, err
	}

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// countDiffLines counts the number of added/removed lines in a diff.
func countDiffLines(diff string) int {
	count := 0
	for _, line := range strings.Split(diff, "\n") {
		if len(line) > 0 && (line[0] == '+' || line[0] == '-') {
			// Skip file headers (--- and +++ lines).
			if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
				continue
			}
			count++
		}
	}
	return count
}

// createRepoSnapshot creates a shallow copy of the repo at the given commit.
// Uses git clone + checkout for isolation (not worktree, to avoid coupling to source).
func createRepoSnapshot(repoPath, commit, destDir string) error {
	// Clone from local repo.
	cmd := exec.Command("git", "clone", "--no-checkout", repoPath, destDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git clone: %s", string(out))
	}

	// Checkout the start commit.
	checkoutCmd := exec.Command("git", "checkout", commit)
	checkoutCmd.Dir = destDir
	if out, err := checkoutCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git checkout %s: %s", commit, string(out))
	}

	return nil
}

// sanitizeTaskName creates a filesystem-safe task name from a commit range.
func sanitizeTaskName(commitRange string) string {
	// Replace ".." with "-to-" and limit to short hashes.
	name := strings.ReplaceAll(commitRange, "..", "-to-")
	// Truncate long hashes to 8 chars each.
	parts := strings.SplitN(name, "-to-", 2)
	if len(parts) == 2 {
		if len(parts[0]) > 8 {
			parts[0] = parts[0][:8]
		}
		if len(parts[1]) > 8 {
			parts[1] = parts[1][:8]
		}
		name = parts[0] + "-to-" + parts[1]
	}
	// Remove any non-alphanumeric/dash characters.
	var sanitized strings.Builder
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' {
			sanitized.WriteRune(c)
		}
	}
	result := sanitized.String()
	if result == "" {
		result = "imported-task"
	}
	return result
}

// tierToMinutes returns a default estimated_minutes for a tier.
func tierToMinutes(tier int) int {
	switch tier {
	case 1:
		return 5
	case 2:
		return 10
	case 3:
		return 20
	case 4:
		return 40
	default:
		return 10
	}
}

// containsFile checks if any file path matches the given name (exact basename match).
func containsFile(files []string, name string) bool {
	for _, f := range files {
		if filepath.Base(f) == name || f == name {
			return true
		}
	}
	return false
}

// containsAnyPrefix checks if any file starts with the given prefix.
func containsAnyPrefix(files []string, prefix string) bool {
	for _, f := range files {
		if strings.HasPrefix(f, prefix) {
			return true
		}
	}
	return false
}

// containsAnySuffix checks if any file ends with the given suffix.
func containsAnySuffix(files []string, suffix string) bool {
	for _, f := range files {
		if strings.HasSuffix(f, suffix) {
			return true
		}
	}
	return false
}

// taskYAMLContent represents the YAML structure for task.yaml.
type taskYAMLContent struct {
	ID               string `yaml:"id"`
	Tier             int    `yaml:"tier"`
	Type             string `yaml:"type"`
	Language         string `yaml:"language"`
	EstimatedMinutes int    `yaml:"estimated_minutes"`
}

// writeTaskYAML writes a task.yaml file.
func writeTaskYAML(path string, content taskYAMLContent) error {
	data, err := yaml.Marshal(content)
	if err != nil {
		return fmt.Errorf("marshal task.yaml: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// writePlanTemplate writes a plan.md template that the user needs to edit.
func writePlanTemplate(path string) error {
	tmpl := `# Task: [EDIT THIS] Brief description

## Goal
[EDIT THIS] What should be accomplished

## Constraints
- [EDIT THIS]

## Acceptance Criteria
- [EDIT THIS]
`
	return os.WriteFile(path, []byte(tmpl), 0644)
}
