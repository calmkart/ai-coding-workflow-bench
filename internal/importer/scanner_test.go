package importer

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectTaskType(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		diff     string
		expected string
	}{
		{
			name:     "reconciler detected",
			files:    []string{"controller.go"},
			diff:     "func (r *Reconciler) Reconcile(ctx context.Context) error {",
			expected: "reconciler",
		},
		{
			name:     "controller-runtime detected",
			files:    []string{"main.go"},
			diff:     `import "sigs.k8s.io/controller-runtime"`,
			expected: "reconciler",
		},
		{
			name:     "http server with main",
			files:    []string{"main.go", "handlers.go"},
			diff:     "http.ListenAndServe(\":8080\", nil)",
			expected: "http-server",
		},
		{
			name:     "gin server with main",
			files:    []string{"cmd/server/main.go", "api.go"},
			diff:     `r := gin.Default()`,
			expected: "http-server",
		},
		{
			name:     "cli with main.go",
			files:    []string{"main.go", "commands.go"},
			diff:     "fmt.Println(result)",
			expected: "cli",
		},
		{
			name:     "cli with cmd/ prefix",
			files:    []string{"cmd/tool/main.go"},
			diff:     "os.Args[1]",
			expected: "cli",
		},
		{
			name:     "library with tests",
			files:    []string{"parser.go", "parser_test.go"},
			diff:     "func Parse(input string) (AST, error) {",
			expected: "library",
		},
		{
			name:     "default to library",
			files:    []string{"utils.go"},
			diff:     "func helper() {}",
			expected: "library",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectTaskType(tt.files, tt.diff)
			if got != tt.expected {
				t.Errorf("detectTaskType(%v, ...) = %q, want %q", tt.files, got, tt.expected)
			}
		})
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected string
	}{
		{"all go files", []string{"main.go", "handler.go", "handler_test.go"}, "go"},
		{"all python files", []string{"app.py", "utils.py", "test_app.py"}, "python"},
		{"all typescript files", []string{"app.ts", "utils.tsx", "index.ts"}, "typescript"},
		{"mixed majority python", []string{"main.go", "app.py", "utils.py", "test.py"}, "python"},
		{"mixed majority typescript", []string{"main.go", "app.ts", "utils.tsx", "index.ts"}, "typescript"},
		{"mixed majority go", []string{"main.go", "handler.go", "utils.go", "app.py"}, "go"},
		{"empty files defaults to go", []string{}, "go"},
		{"no recognized extensions defaults to go", []string{"README.md", "Makefile"}, "go"},
		{"tie go and python defaults to go", []string{"main.go", "app.py"}, "go"},
		{"tie go and ts defaults to go", []string{"main.go", "app.ts"}, "go"},
		{"tie python and ts defaults to go", []string{"app.py", "app.ts"}, "go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectLanguage(tt.files)
			if got != tt.expected {
				t.Errorf("detectLanguage(%v) = %q, want %q", tt.files, got, tt.expected)
			}
		})
	}
}

func TestEstimateTier(t *testing.T) {
	tests := []struct {
		name      string
		diffLines int
		fileCount int
		expected  int
	}{
		{"small change", 20, 1, 1},
		{"small files max", 49, 2, 1},
		{"tier2 boundary", 50, 2, 2},
		{"tier2 mid", 100, 3, 2},
		{"tier2 files max", 150, 5, 2},
		{"tier3 boundary", 200, 3, 3},
		{"tier3 mid", 350, 8, 3},
		{"tier3 max", 499, 10, 3},
		{"tier4 large", 500, 12, 4},
		{"tier4 very large", 2000, 50, 4},
		// Edge: small diff but many files => tier2 due to file count
		{"small diff many files", 30, 5, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := estimateTier(tt.diffLines, tt.fileCount)
			if got != tt.expected {
				t.Errorf("estimateTier(%d, %d) = %d, want %d", tt.diffLines, tt.fileCount, got, tt.expected)
			}
		})
	}
}

func TestCountDiffLines(t *testing.T) {
	diff := `diff --git a/foo.go b/foo.go
--- a/foo.go
+++ b/foo.go
@@ -1,3 +1,4 @@
 package foo
+import "fmt"
-var x = 1
+var x = 2
+var y = 3
`
	got := countDiffLines(diff)
	// +import "fmt", -var x = 1, +var x = 2, +var y = 3 = 4
	if got != 4 {
		t.Errorf("countDiffLines() = %d, want 4", got)
	}
}

func TestSanitizeTaskName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"abc123..def456", "abc123-to-def456"},
		{"abcdef0123456789..0123456789abcdef", "abcdef01-to-01234567"},
		{"a..b", "a-to-b"},
		{"no-dots-here", "no-dots-here"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeTaskName(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeTaskName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestContainsFile(t *testing.T) {
	files := []string{"cmd/main.go", "handlers.go", "internal/store/db.go"}
	if !containsFile(files, "main.go") {
		t.Error("expected containsFile to find main.go by basename")
	}
	if !containsFile(files, "handlers.go") {
		t.Error("expected containsFile to find handlers.go")
	}
	if containsFile(files, "nonexistent.go") {
		t.Error("expected containsFile not to find nonexistent.go")
	}
}

func TestContainsAnyPrefix(t *testing.T) {
	files := []string{"cmd/main.go", "handlers.go"}
	if !containsAnyPrefix(files, "cmd/") {
		t.Error("expected containsAnyPrefix to find cmd/ prefix")
	}
	if containsAnyPrefix(files, "internal/") {
		t.Error("expected containsAnyPrefix not to find internal/ prefix")
	}
}

func TestContainsAnySuffix(t *testing.T) {
	files := []string{"foo.go", "bar_test.go"}
	if !containsAnySuffix(files, "_test.go") {
		t.Error("expected containsAnySuffix to find _test.go suffix")
	}
	if containsAnySuffix(files, "_bench.go") {
		t.Error("expected containsAnySuffix not to find _bench.go suffix")
	}
}

func TestTierToMinutes(t *testing.T) {
	tests := []struct {
		tier     int
		expected int
	}{
		{1, 5},
		{2, 10},
		{3, 20},
		{4, 40},
		{0, 10},
		{5, 10},
	}
	for _, tt := range tests {
		got := tierToMinutes(tt.tier)
		if got != tt.expected {
			t.Errorf("tierToMinutes(%d) = %d, want %d", tt.tier, got, tt.expected)
		}
	}
}

func TestImport_EmptyRepo(t *testing.T) {
	_, err := Import(ImportConfig{RepoPath: "", CommitRange: "a..b"})
	if err == nil {
		t.Fatal("expected error for empty repo path")
	}
}

func TestImport_EmptyCommitRange(t *testing.T) {
	_, err := Import(ImportConfig{RepoPath: "/tmp", CommitRange: ""})
	if err == nil {
		t.Fatal("expected error for empty commit range")
	}
}

func TestImport_InvalidCommitRange(t *testing.T) {
	// Create a temp git repo.
	tmpDir := t.TempDir()
	initGitRepo(t, tmpDir)

	_, err := Import(ImportConfig{RepoPath: tmpDir, CommitRange: "invalid-range"})
	if err == nil {
		t.Fatal("expected error for invalid commit range format")
	}
}

func TestImport_NotGitRepo(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := Import(ImportConfig{RepoPath: tmpDir, CommitRange: "a..b"})
	if err == nil {
		t.Fatal("expected error for non-git directory")
	}
}

func TestImport_FullFlow(t *testing.T) {
	// Check that git is available.
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	// Create a temp git repo with two commits.
	repoDir := t.TempDir()
	initGitRepo(t, repoDir)

	// Create initial commit.
	writeFile(t, filepath.Join(repoDir, "main.go"), `package main
func main() {}
`)
	gitAdd(t, repoDir, "main.go")
	commit1 := gitCommit(t, repoDir, "initial commit")

	// Create second commit with HTTP server code.
	writeFile(t, filepath.Join(repoDir, "main.go"), `package main

import "net/http"

func main() {
	http.ListenAndServe(":8080", nil)
}
`)
	writeFile(t, filepath.Join(repoDir, "handlers.go"), `package main

import "net/http"

func handleTodos(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("todos"))
}
`)
	gitAdd(t, repoDir, "main.go")
	gitAdd(t, repoDir, "handlers.go")
	commit2 := gitCommit(t, repoDir, "add http server")

	// Import the task.
	outputDir := filepath.Join(t.TempDir(), "output")
	result, err := Import(ImportConfig{
		RepoPath:    repoDir,
		CommitRange: commit1 + ".." + commit2,
		OutputDir:   outputDir,
	})
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	// Verify result.
	if result.Type != "http-server" {
		t.Errorf("expected type 'http-server', got %q", result.Type)
	}
	if result.Tier < 1 || result.Tier > 4 {
		t.Errorf("expected tier 1-4, got %d", result.Tier)
	}
	if result.DiffSize <= 0 {
		t.Errorf("expected positive diff size, got %d", result.DiffSize)
	}
	if len(result.Files) == 0 {
		t.Error("expected non-empty files list")
	}

	// Verify directory structure.
	if _, err := os.Stat(filepath.Join(outputDir, "task.yaml")); err != nil {
		t.Errorf("task.yaml not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "plan.md")); err != nil {
		t.Errorf("plan.md not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "repo")); err != nil {
		t.Errorf("repo/ not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, "verify")); err != nil {
		t.Errorf("verify/ not created: %v", err)
	}

	// Verify plan.md contains the template marker.
	planData, err := os.ReadFile(filepath.Join(outputDir, "plan.md"))
	if err != nil {
		t.Fatalf("read plan.md: %v", err)
	}
	if !strings.Contains(string(planData), "[EDIT THIS]") {
		t.Error("plan.md should contain [EDIT THIS] template markers")
	}
}

// Helper functions for git operations in tests.

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	run(t, dir, "git", "init")
	run(t, dir, "git", "config", "user.email", "test@example.com")
	run(t, dir, "git", "config", "user.name", "Test")
}

func gitAdd(t *testing.T, dir, file string) {
	t.Helper()
	run(t, dir, "git", "add", file)
}

func gitCommit(t *testing.T, dir, msg string) string {
	t.Helper()
	run(t, dir, "git", "commit", "-m", msg)
	out := runOutput(t, dir, "git", "rev-parse", "HEAD")
	return out
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, string(out))
	}
}

func runOutput(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s %v failed: %v", name, args, err)
	}
	return strings.TrimSpace(string(out))
}

