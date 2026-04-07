package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseVerifyOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantL1   bool
		wantL2P  int
		wantL2T  int
		wantL3   int
		wantL4P  int
		wantL4T  int
		wantErr  bool
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
