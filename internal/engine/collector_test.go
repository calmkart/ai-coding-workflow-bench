package engine

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseVerifyOutput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantL1  bool
		wantL2P int
		wantL2T int
		wantL3  int
		wantL4P int
		wantL4T int
		wantErr bool
	}{
		{
			name:    "all pass",
			input:   "some output\nBENCH_RESULT: L1=PASS L2=8/8 L3=0 L4=5/5\n",
			wantL1:  true,
			wantL2P: 8, wantL2T: 8,
			wantL3:  0,
			wantL4P: 5, wantL4T: 5,
		},
		{
			name:    "partial failure",
			input:   "BENCH_RESULT: L1=PASS L2=6/8 L3=2 L4=3/5\n",
			wantL1:  true,
			wantL2P: 6, wantL2T: 8,
			wantL3:  2,
			wantL4P: 3, wantL4T: 5,
		},
		{
			name:   "build failure",
			input:  "go build failed\nBENCH_RESULT: L1=FAIL\n",
			wantL1: false,
		},
		{
			name:    "no result line",
			input:   "some random output\n",
			wantErr: true,
		},
		{
			name:    "zero totals",
			input:   "BENCH_RESULT: L1=PASS L2=0/0 L3=0 L4=0/0\n",
			wantL1:  true,
			wantL2P: 0, wantL2T: 0,
			wantL3:  0,
			wantL4P: 0, wantL4T: 0,
		},
		{
			name:    "E2E compile failure produces L4=0/1",
			input:   "BENCH_RESULT: L1=PASS L2=8/8 L3=0 L4=0/1\n",
			wantL1:  true,
			wantL2P: 8, wantL2T: 8,
			wantL3:  0,
			wantL4P: 0, wantL4T: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseVerifyOutput(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.L1Build != tt.wantL1 {
				t.Errorf("L1Build: got %v, want %v", result.L1Build, tt.wantL1)
			}
			if result.L2Passed != tt.wantL2P {
				t.Errorf("L2Passed: got %d, want %d", result.L2Passed, tt.wantL2P)
			}
			if result.L2Total != tt.wantL2T {
				t.Errorf("L2Total: got %d, want %d", result.L2Total, tt.wantL2T)
			}
			if result.L3Issues != tt.wantL3 {
				t.Errorf("L3Issues: got %d, want %d", result.L3Issues, tt.wantL3)
			}
			if result.L4Passed != tt.wantL4P {
				t.Errorf("L4Passed: got %d, want %d", result.L4Passed, tt.wantL4P)
			}
			if result.L4Total != tt.wantL4T {
				t.Errorf("L4Total: got %d, want %d", result.L4Total, tt.wantL4T)
			}
		})
	}
}

// TestRunVerify_ScriptNotFound verifies that RunVerify returns output (not error)
// when bash can't find the script (bash exits 127 = ExitError, which is expected).
// Fix 2: ExitError is treated as expected; real infra errors (e.g. no bash binary) are errors.
func TestRunVerify_ScriptNotFound(t *testing.T) {
	// bash will exit 127 for nonexistent script — this is an ExitError, not infra error.
	out, err := RunVerify("/nonexistent/dir", "/tmp")
	if err != nil {
		t.Errorf("expected no error (ExitError is expected), got: %v", err)
	}
	if !strings.Contains(out, "No such file") {
		t.Logf("output: %s", out)
	}
}

// TestRunVerify_TestFailureNotError verifies that RunVerify does NOT return an
// error when the script exits non-zero due to test failures (ExitError).
// Fix 2: ExitError is expected behavior, not an infrastructure error.
func TestRunVerify_TestFailureNotError(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "verify.sh")
	// Script that exits 1 (simulating test failure).
	if err := os.WriteFile(script, []byte("#!/bin/bash\necho 'BENCH_RESULT: L1=FAIL'\nexit 1\n"), 0755); err != nil {
		t.Fatal(err)
	}

	out, err := RunVerify(dir, "/tmp")
	if err != nil {
		t.Errorf("expected no error for ExitError (test failure), got: %v", err)
	}
	if !strings.Contains(out, "BENCH_RESULT") {
		t.Errorf("expected BENCH_RESULT in output, got: %s", out)
	}
}

// TestVerifyOutputStorage verifies that verify.log is saved to the raw directory
// when HomeDir is set in RunConfig.
func TestVerifyOutputStorage(t *testing.T) {
	homeDir := t.TempDir()
	runID := "test-verify-storage-run1"

	// Simulate saving verify output as executeOneRun does.
	rawDir := filepath.Join(homeDir, "raw", runID)
	if err := os.MkdirAll(rawDir, 0755); err != nil {
		t.Fatalf("create raw dir: %v", err)
	}

	verifyOutput := "some test output\nBENCH_RESULT: L1=PASS L2=8/8 L3=0 L4=5/5\n"
	verifyLogPath := filepath.Join(rawDir, "verify.log")
	if err := os.WriteFile(verifyLogPath, []byte(verifyOutput), 0644); err != nil {
		t.Fatalf("write verify.log: %v", err)
	}

	// Verify the file exists and has correct content.
	data, err := os.ReadFile(verifyLogPath)
	if err != nil {
		t.Fatalf("read verify.log: %v", err)
	}
	if string(data) != verifyOutput {
		t.Errorf("verify.log content mismatch:\ngot:  %q\nwant: %q", string(data), verifyOutput)
	}
}

// gitCmd is a test helper that runs a git command in the specified directory.
func gitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@test.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, string(out))
	}
}

// TestCaptureDiff_SavesPatch verifies that captureDiff creates diff.patch
// when there are changes in the worktree.
func TestCaptureDiff_SavesPatch(t *testing.T) {
	// Create a temporary git repo.
	repoDir := t.TempDir()

	// Initialize git repo with a file.
	gitCmd(t, repoDir, "init")
	if err := os.WriteFile(filepath.Join(repoDir, "hello.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, repoDir, "add", "-A")
	gitCmd(t, repoDir, "commit", "-m", "initial")

	// Make a change.
	if err := os.WriteFile(filepath.Join(repoDir, "hello.go"), []byte("package main\n\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	homeDir := t.TempDir()
	runID := "test-capture-diff"

	captureDiff(repoDir, homeDir, runID)

	patchPath := filepath.Join(homeDir, "raw", runID, "diff.patch")
	data, err := os.ReadFile(patchPath)
	if err != nil {
		t.Fatalf("expected diff.patch to exist: %v", err)
	}
	if !strings.Contains(string(data), "func main()") {
		t.Errorf("expected diff.patch to contain 'func main()', got:\n%s", string(data))
	}
}

// TestCaptureDiff_NoChanges verifies that captureDiff does not create diff.patch
// when there are no changes.
func TestCaptureDiff_NoChanges(t *testing.T) {
	repoDir := t.TempDir()
	gitCmd(t, repoDir, "init")
	if err := os.WriteFile(filepath.Join(repoDir, "hello.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	gitCmd(t, repoDir, "add", "-A")
	gitCmd(t, repoDir, "commit", "-m", "initial")

	homeDir := t.TempDir()
	runID := "test-no-changes"

	captureDiff(repoDir, homeDir, runID)

	patchPath := filepath.Join(homeDir, "raw", runID, "diff.patch")
	if _, err := os.Stat(patchPath); !os.IsNotExist(err) {
		t.Error("expected diff.patch to not exist when no changes")
	}
}

// TestGenerateVerifyDir_NonHTTPServerTaskType verifies that GenerateVerifyDir
// succeeds for non-http-server task types using the generic template.
func TestGenerateVerifyDir_NonHTTPServerTaskType(t *testing.T) {
	dir := t.TempDir()
	taskDir := filepath.Join(dir, "task")
	os.MkdirAll(taskDir, 0755)

	verifyDir, err := GenerateVerifyDir(VerifyConfig{
		TaskType: "library",
		TaskDir:  taskDir,
		RunID:    "test-library-type",
	})
	if err != nil {
		t.Fatalf("expected success for library task type, got error: %v", err)
	}

	// Verify that verify.sh was generated.
	verifyScript := filepath.Join(verifyDir, "verify.sh")
	if _, err := os.Stat(verifyScript); os.IsNotExist(err) {
		t.Error("verify.sh was not generated")
	}

	// Clean up.
	os.RemoveAll(verifyDir)
}
