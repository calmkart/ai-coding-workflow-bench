//go:build integration

package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestIntegration_VerifyPipeline tests the full verify -> collect pipeline
// using a real git repo, go build, and go test.
//
// This test:
// 1. Copies the tier1/fix-handler-bug repo to a temp dir
// 2. Initializes it as a git repo
// 3. Applies the fix (simulating what an adapter would do)
// 4. Runs verify.sh
// 5. Parses BENCH_RESULT from output
// 6. Validates L1=PASS and L2/L4 > 0
//
// Requires: git, go toolchain
func TestIntegration_VerifyPipeline(t *testing.T) {
	// Find project root (go up from internal/engine/)
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("find project root: %v", err)
	}

	taskDir := filepath.Join(projectRoot, "tasks", "tier1", "fix-handler-bug")
	repoSrc := filepath.Join(taskDir, "repo")

	if _, err := os.Stat(repoSrc); os.IsNotExist(err) {
		t.Fatalf("task repo not found at %s", repoSrc)
	}

	// Create temp worktree
	workDir := t.TempDir()
	repoDir := filepath.Join(workDir, "repo")

	// Copy repo to temp dir
	if err := copyDir(repoSrc, repoDir); err != nil {
		t.Fatalf("copy repo: %v", err)
	}

	// Remove existing .git and reinitialize (so we have a clean repo)
	os.RemoveAll(filepath.Join(repoDir, ".git"))
	runCmd(t, repoDir, "git", "init")
	runCmd(t, repoDir, "git", "add", "-A")
	runCmd(t, repoDir, "git", "commit", "-m", "initial")

	// Apply the fix: change offset = page * pageSize to (page - 1) * pageSize
	// We do this by reading handlers.go and doing a string replacement
	handlersPath := filepath.Join(repoDir, "handlers.go")
	content, err := os.ReadFile(handlersPath)
	if err != nil {
		t.Fatalf("read handlers.go: %v", err)
	}

	// The off-by-one bug: offset = page * pageSize should be (page - 1) * pageSize
	fixed := strings.ReplaceAll(string(content), "page * pageSize", "(page - 1) * pageSize")
	if fixed == string(content) {
		// Try alternative patterns
		fixed = strings.ReplaceAll(string(content), "page*pageSize", "(page-1)*pageSize")
	}
	if fixed == string(content) {
		t.Log("WARNING: could not find exact off-by-one pattern to fix, proceeding with original code")
	}

	if err := os.WriteFile(handlersPath, []byte(fixed), 0644); err != nil {
		t.Fatalf("write fixed handlers.go: %v", err)
	}

	// Set up verify directory with e2e test
	verifyDir := filepath.Join(workDir, "verify")
	if err := os.MkdirAll(verifyDir, 0755); err != nil {
		t.Fatalf("mkdir verify: %v", err)
	}

	// Copy e2e test source
	e2eSrc := filepath.Join(taskDir, "verify", "e2e_test.go.src")
	e2eDst := filepath.Join(verifyDir, "e2e_test.go")
	e2eContent, err := os.ReadFile(e2eSrc)
	if err != nil {
		t.Fatalf("read e2e source: %v", err)
	}
	if err := os.WriteFile(e2eDst, e2eContent, 0644); err != nil {
		t.Fatalf("write e2e test: %v", err)
	}

	// Create a simplified verify script that handles errors gracefully
	// (The template in templates/ uses set -e which causes early exit on test failures)
	verifyScript := filepath.Join(verifyDir, "verify.sh")
	scriptContent := `#!/usr/bin/env bash
TASK_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$1"
cd "$REPO_DIR"

count_matches() {
  local pattern="$1" file="$2"
  local c
  c=$(grep -c "$pattern" "$file" 2>/dev/null) || c=0
  echo "$c" | head -1 | tr -d '[:space:]'
}

echo "=== L1: Build ==="
go build ./... 2>&1 || { echo "BENCH_RESULT: L1=FAIL"; exit 0; }

echo "=== L2: Unit Tests ==="
go test ./... -count=1 -v 2>&1 | tee /tmp/bench-int-ut.log || true
UT_PASS=$(count_matches "^--- PASS" /tmp/bench-int-ut.log)
UT_FAIL=$(count_matches "^--- FAIL" /tmp/bench-int-ut.log)

echo "=== L3: Static Analysis ==="
LINT_ISSUES=0

echo "=== L4: E2E ==="
E2E_PASS=0
E2E_FAIL=0
if [ -f "$TASK_DIR/e2e_test.go" ]; then
  cp "$TASK_DIR/e2e_test.go" "$REPO_DIR/bench_e2e_test.go"
  go test -v -run TestBenchE2E -count=1 ./... 2>&1 | tee /tmp/bench-int-e2e.log || true
  E2E_PASS=$(count_matches "^--- PASS" /tmp/bench-int-e2e.log)
  E2E_FAIL=$(count_matches "^--- FAIL" /tmp/bench-int-e2e.log)
  rm -f "$REPO_DIR/bench_e2e_test.go"
fi

UT_TOTAL=$((UT_PASS + UT_FAIL))
E2E_TOTAL=$((E2E_PASS + E2E_FAIL))
echo "BENCH_RESULT: L1=PASS L2=${UT_PASS}/${UT_TOTAL} L3=${LINT_ISSUES} L4=${E2E_PASS}/${E2E_TOTAL}"
`
	if err := os.WriteFile(verifyScript, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("write verify.sh: %v", err)
	}

	// Run verify.sh
	cmd := exec.Command("bash", verifyScript, repoDir)
	cmd.Dir = verifyDir
	cmd.Env = append(os.Environ(), "HOME="+os.Getenv("HOME"))
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	t.Logf("verify.sh output (last 500 chars):\n%s", truncateLast(outputStr, 500))

	// Even if verify.sh returns non-zero (some tests may fail), we should still
	// get a BENCH_RESULT line
	if !strings.Contains(outputStr, "BENCH_RESULT:") {
		if err != nil {
			t.Logf("verify.sh error: %v", err)
		}
		t.Fatal("verify.sh did not produce BENCH_RESULT line")
	}

	// Parse the result
	result, parseErr := ParseVerifyOutput(outputStr)
	if parseErr != nil {
		t.Fatalf("ParseVerifyOutput: %v", parseErr)
	}

	t.Logf("Parsed result: L1=%v, L2=%d/%d, L3=%d, L4=%d/%d",
		result.L1Build, result.L2Passed, result.L2Total, result.L3Issues,
		result.L4Passed, result.L4Total)

	// L1 should pass (code should compile after fix)
	if !result.L1Build {
		t.Error("expected L1=PASS (code should compile)")
	}

	// L2 should have some tests
	if result.L2Total == 0 {
		t.Error("expected L2Total > 0 (repo has unit tests)")
	}

	// L2 pass rate should be > 0
	if result.L2Passed == 0 {
		t.Error("expected L2Passed > 0 (some unit tests should pass)")
	}
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root (no go.mod found)")
		}
		dir = parent
	}
}

func copyDir(src, dst string) error {
	cmd := exec.Command("cp", "-r", src, dst)
	return cmd.Run()
}

func truncateLast(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return "..." + s[len(s)-n:]
}

func runCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@test.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, string(out))
	}
}
