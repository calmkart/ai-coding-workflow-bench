package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGoVerifyTemplate_Embedded verifies the renamed template is properly embedded.
func TestGoVerifyTemplate_Embedded(t *testing.T) {
	if goVerifyTemplate == "" {
		t.Fatal("goVerifyTemplate is empty -- embed failed")
	}
	// The template should contain key verify script markers.
	if !strings.Contains(goVerifyTemplate, "BENCH_RESULT") {
		t.Error("expected BENCH_RESULT in verify template")
	}
	if !strings.Contains(goVerifyTemplate, "L1: Build") {
		t.Error("expected L1: Build in verify template")
	}
	if !strings.Contains(goVerifyTemplate, "L4: E2E") {
		t.Error("expected L4: E2E in verify template")
	}
}

// TestGoVerifyTemplate_UsesJSONOutput verifies the template uses go test -json
// for accurate test counting (P7).
func TestGoVerifyTemplate_UsesJSONOutput(t *testing.T) {
	// L2 section should use go test -json instead of grep "--- PASS".
	if !strings.Contains(goVerifyTemplate, "go test -json ./... -count=1 -race") {
		t.Error("expected 'go test -json' in L2 section of verify template")
	}
	// Should NOT use the old grep "--- PASS" method.
	if strings.Contains(goVerifyTemplate, `grep -c -- "--- PASS"`) {
		t.Error("verify template should not use grep '--- PASS' anymore (P7)")
	}
	// L4 E2E section should also use go test -json.
	if !strings.Contains(goVerifyTemplate, "go test -json -run TestBenchE2E") {
		t.Error("expected 'go test -json -run TestBenchE2E' in L4 section of verify template")
	}
	// Should filter top-level tests (no subtest '/').
	if !strings.Contains(goVerifyTemplate, `grep -v '"Test":"[^"]*/'`) {
		t.Error("expected subtest filtering in verify template")
	}
}

// TestGenerateVerifyDir_CreatesFiles verifies that GenerateVerifyDir creates
// the expected verify.sh file from the renamed template.
func TestGenerateVerifyDir_CreatesFiles(t *testing.T) {
	// Create a minimal task dir with verify/ directory.
	taskDir := t.TempDir()
	verifyTaskDir := filepath.Join(taskDir, "verify")
	if err := os.MkdirAll(verifyTaskDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create a dummy e2e_test.go.src.
	e2eSrc := filepath.Join(verifyTaskDir, "e2e_test.go.src")
	if err := os.WriteFile(e2eSrc, []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	dir, err := GenerateVerifyDir(VerifyConfig{
		TaskType: "http-server",
		TaskDir:  taskDir,
		RunID:    "test-verify-rename",
	})
	if err != nil {
		t.Fatalf("GenerateVerifyDir: %v", err)
	}
	defer os.RemoveAll(dir)

	// verify.sh should exist.
	verifyPath := filepath.Join(dir, "verify.sh")
	data, err := os.ReadFile(verifyPath)
	if err != nil {
		t.Fatalf("read verify.sh: %v", err)
	}
	if !strings.Contains(string(data), "BENCH_RESULT") {
		t.Error("verify.sh missing BENCH_RESULT line")
	}

	// e2e_test.go should be copied.
	e2eDst := filepath.Join(dir, "e2e_test.go")
	if _, err := os.Stat(e2eDst); os.IsNotExist(err) {
		t.Error("e2e_test.go was not copied to verify dir")
	}
}

// --- P18: Python + TypeScript Verify Template Tests ---

// TestPythonVerifyTemplate_Embedded verifies the Python template is properly embedded.
func TestPythonVerifyTemplate_Embedded(t *testing.T) {
	if pythonVerifyTemplate == "" {
		t.Fatal("pythonVerifyTemplate is empty -- embed failed")
	}
	if !strings.Contains(pythonVerifyTemplate, "BENCH_RESULT") {
		t.Error("expected BENCH_RESULT in python verify template")
	}
	if !strings.Contains(pythonVerifyTemplate, "L1: Syntax Check") {
		t.Error("expected L1: Syntax Check in python verify template")
	}
	if !strings.Contains(pythonVerifyTemplate, "L4: E2E") {
		t.Error("expected L4: E2E in python verify template")
	}
	if !strings.Contains(pythonVerifyTemplate, "py_compile") {
		t.Error("expected py_compile in python verify template")
	}
	if !strings.Contains(pythonVerifyTemplate, "pytest") {
		t.Error("expected pytest in python verify template")
	}
}

// TestTypescriptVerifyTemplate_Embedded verifies the TypeScript template is properly embedded.
func TestTypescriptVerifyTemplate_Embedded(t *testing.T) {
	if typescriptVerifyTemplate == "" {
		t.Fatal("typescriptVerifyTemplate is empty -- embed failed")
	}
	if !strings.Contains(typescriptVerifyTemplate, "BENCH_RESULT") {
		t.Error("expected BENCH_RESULT in typescript verify template")
	}
	if !strings.Contains(typescriptVerifyTemplate, "L1: Type Check") {
		t.Error("expected L1: Type Check in typescript verify template")
	}
	if !strings.Contains(typescriptVerifyTemplate, "L4: E2E") {
		t.Error("expected L4: E2E in typescript verify template")
	}
	if !strings.Contains(typescriptVerifyTemplate, "tsc --noEmit") {
		t.Error("expected tsc --noEmit in typescript verify template")
	}
	if !strings.Contains(typescriptVerifyTemplate, "vitest") {
		t.Error("expected vitest in typescript verify template")
	}
	if !strings.Contains(typescriptVerifyTemplate, "eslint") {
		t.Error("expected eslint in typescript verify template")
	}
}

// TestGenerateVerifyDir_PythonLanguage verifies that GenerateVerifyDir selects
// the Python template and E2E file when TaskLanguage is "python".
func TestGenerateVerifyDir_PythonLanguage(t *testing.T) {
	taskDir := t.TempDir()
	verifyTaskDir := filepath.Join(taskDir, "verify")
	if err := os.MkdirAll(verifyTaskDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create a Python e2e test file.
	e2eSrc := filepath.Join(verifyTaskDir, "e2e_test.py")
	if err := os.WriteFile(e2eSrc, []byte("def test_e2e(): pass\n"), 0644); err != nil {
		t.Fatal(err)
	}

	dir, err := GenerateVerifyDir(VerifyConfig{
		TaskType:     "library",
		TaskDir:      taskDir,
		TaskLanguage: "python",
		RunID:        "test-python-verify",
	})
	if err != nil {
		t.Fatalf("GenerateVerifyDir: %v", err)
	}
	defer os.RemoveAll(dir)

	// verify.sh should use Python template (contains py_compile).
	verifyPath := filepath.Join(dir, "verify.sh")
	data, err := os.ReadFile(verifyPath)
	if err != nil {
		t.Fatalf("read verify.sh: %v", err)
	}
	if !strings.Contains(string(data), "py_compile") {
		t.Error("verify.sh should contain py_compile for python language")
	}
	if strings.Contains(string(data), "go build") {
		t.Error("verify.sh should not contain 'go build' for python language")
	}

	// E2E file should be copied as e2e_test.py.
	e2eDst := filepath.Join(dir, "e2e_test.py")
	if _, err := os.Stat(e2eDst); os.IsNotExist(err) {
		t.Error("e2e_test.py was not copied to verify dir")
	}
}

// TestGenerateVerifyDir_TypescriptLanguage verifies that GenerateVerifyDir selects
// the TypeScript template and E2E file when TaskLanguage is "typescript".
func TestGenerateVerifyDir_TypescriptLanguage(t *testing.T) {
	taskDir := t.TempDir()
	verifyTaskDir := filepath.Join(taskDir, "verify")
	if err := os.MkdirAll(verifyTaskDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create a TypeScript e2e test file with .src extension.
	e2eSrc := filepath.Join(verifyTaskDir, "e2e_test.ts.src")
	if err := os.WriteFile(e2eSrc, []byte("test('e2e', () => {})\n"), 0644); err != nil {
		t.Fatal(err)
	}

	dir, err := GenerateVerifyDir(VerifyConfig{
		TaskType:     "library",
		TaskDir:      taskDir,
		TaskLanguage: "typescript",
		RunID:        "test-typescript-verify",
	})
	if err != nil {
		t.Fatalf("GenerateVerifyDir: %v", err)
	}
	defer os.RemoveAll(dir)

	// verify.sh should use TypeScript template (contains tsc).
	verifyPath := filepath.Join(dir, "verify.sh")
	data, err := os.ReadFile(verifyPath)
	if err != nil {
		t.Fatalf("read verify.sh: %v", err)
	}
	if !strings.Contains(string(data), "tsc --noEmit") {
		t.Error("verify.sh should contain tsc --noEmit for typescript language")
	}
	if strings.Contains(string(data), "go build") {
		t.Error("verify.sh should not contain 'go build' for typescript language")
	}

	// E2E file should be copied as e2e_test.ts (without .src).
	e2eDst := filepath.Join(dir, "e2e_test.ts")
	if _, err := os.Stat(e2eDst); os.IsNotExist(err) {
		t.Error("e2e_test.ts was not copied to verify dir")
	}
}

// TestGenerateVerifyDir_DefaultLanguageIsGo verifies that empty TaskLanguage
// defaults to Go template (backwards compatible).
func TestGenerateVerifyDir_DefaultLanguageIsGo(t *testing.T) {
	taskDir := t.TempDir()
	os.MkdirAll(filepath.Join(taskDir, "verify"), 0755)

	dir, err := GenerateVerifyDir(VerifyConfig{
		TaskType:     "library",
		TaskDir:      taskDir,
		TaskLanguage: "", // empty -> default to Go
		RunID:        "test-default-go",
	})
	if err != nil {
		t.Fatalf("GenerateVerifyDir: %v", err)
	}
	defer os.RemoveAll(dir)

	data, err := os.ReadFile(filepath.Join(dir, "verify.sh"))
	if err != nil {
		t.Fatalf("read verify.sh: %v", err)
	}
	if !strings.Contains(string(data), "go build") {
		t.Error("default language should produce Go verify template with 'go build'")
	}
}

// TestGenerateVerifyDir_UnknownLanguageFallsBackToGo verifies that an unknown
// language falls back to the Go template.
func TestGenerateVerifyDir_UnknownLanguageFallsBackToGo(t *testing.T) {
	taskDir := t.TempDir()
	os.MkdirAll(filepath.Join(taskDir, "verify"), 0755)

	dir, err := GenerateVerifyDir(VerifyConfig{
		TaskType:     "library",
		TaskDir:      taskDir,
		TaskLanguage: "rust", // unknown -> fallback to Go
		RunID:        "test-unknown-lang",
	})
	if err != nil {
		t.Fatalf("GenerateVerifyDir: %v", err)
	}
	defer os.RemoveAll(dir)

	data, err := os.ReadFile(filepath.Join(dir, "verify.sh"))
	if err != nil {
		t.Fatalf("read verify.sh: %v", err)
	}
	if !strings.Contains(string(data), "go build") {
		t.Error("unknown language should fall back to Go verify template")
	}
}

// TestE2EFileForLanguage verifies E2E file naming for each language.
func TestE2EFileForLanguage(t *testing.T) {
	tests := []struct {
		lang    string
		wantSrc string
		wantDst string
	}{
		{"go", "e2e_test.go.src", "e2e_test.go"},
		{"", "e2e_test.go.src", "e2e_test.go"},
		{"python", "e2e_test.py", "e2e_test.py"},
		{"typescript", "e2e_test.ts.src", "e2e_test.ts"},
	}
	for _, tt := range tests {
		t.Run("lang="+tt.lang, func(t *testing.T) {
			cfg := e2eFileForLanguage(tt.lang)
			if cfg.srcName != tt.wantSrc {
				t.Errorf("srcName: got %q, want %q", cfg.srcName, tt.wantSrc)
			}
			if cfg.dstName != tt.wantDst {
				t.Errorf("dstName: got %q, want %q", cfg.dstName, tt.wantDst)
			}
		})
	}
}
