package engine

import (
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

//go:embed templates/go_verify.sh.tmpl
var goVerifyTemplate string

//go:embed templates/python_verify.sh.tmpl
var pythonVerifyTemplate string

//go:embed templates/typescript_verify.sh.tmpl
var typescriptVerifyTemplate string

// VerifyConfig holds parameters for generating a verify script.
type VerifyConfig struct {
	TaskType     string // http-server, k8s-operator, library, cli
	TaskDir      string // path to task directory (contains verify/e2e_test.go)
	TaskLanguage string // "go", "python", "typescript"; defaults to "go" if empty
	RunID        string
}

// e2eFileConfig maps task languages to their E2E test source/destination filenames.
// Go uses .src extension to avoid Go tooling compiling it in place.
// Python files can use .py directly (no auto-compilation concern).
// TypeScript uses .src extension like Go.
type e2eFileConfig struct {
	srcName string // filename in verify/ directory
	dstName string // filename copied into verify temp dir
}

// e2eFileForLanguage returns the E2E source and destination filenames for the given language.
//
// @implements P18 (E2E test file extension mapping per language)
func e2eFileForLanguage(lang string) e2eFileConfig {
	switch lang {
	case "python":
		return e2eFileConfig{srcName: "e2e_test.py", dstName: "e2e_test.py"}
	case "typescript":
		return e2eFileConfig{srcName: "e2e_test.ts.src", dstName: "e2e_test.ts"}
	default: // "go" or empty
		return e2eFileConfig{srcName: "e2e_test.go.src", dstName: "e2e_test.go"}
	}
}

// GenerateVerifyDir creates a temporary directory with verify.sh and the E2E test file.
// Returns the path to the temporary directory. The verify template and E2E file extension
// are selected based on TaskLanguage.
//
// @implements REQ-VERIFY-GEN (generate verify.sh from template + copy e2e test)
// @implements P18 (Python + TypeScript verify templates)
func GenerateVerifyDir(cfg VerifyConfig) (string, error) {
	verifyDir := filepath.Join(os.TempDir(), "bench-verify-"+cfg.RunID)
	if err := os.MkdirAll(verifyDir, 0755); err != nil {
		return "", fmt.Errorf("create verify dir: %w", err)
	}

	// Select template based on task language.
	var tmplStr string
	switch cfg.TaskLanguage {
	case "python":
		tmplStr = pythonVerifyTemplate
	case "typescript":
		tmplStr = typescriptVerifyTemplate
	case "go", "":
		tmplStr = goVerifyTemplate
	default:
		slog.Info("unknown language, using go verify template", "language", cfg.TaskLanguage)
		tmplStr = goVerifyTemplate
	}

	// Generate verify.sh from template.
	tmpl, err := template.New("verify").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("parse verify template: %w", err)
	}

	verifyPath := filepath.Join(verifyDir, "verify.sh")
	f, err := os.OpenFile(verifyPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return "", fmt.Errorf("create verify.sh: %w", err)
	}
	defer f.Close()

	// Fix 4: Pass empty struct instead of nil to template Execute.
	if err := tmpl.Execute(f, struct{}{}); err != nil {
		return "", fmt.Errorf("execute verify template: %w", err)
	}

	// Copy E2E test file from task's verify/ directory.
	e2eCfg := e2eFileForLanguage(cfg.TaskLanguage)
	e2eSrc := filepath.Join(cfg.TaskDir, "verify", e2eCfg.srcName)
	e2eDst := filepath.Join(verifyDir, e2eCfg.dstName)
	if data, err := os.ReadFile(e2eSrc); err == nil {
		if err := os.WriteFile(e2eDst, data, 0644); err != nil {
			return "", fmt.Errorf("copy %s: %w", e2eCfg.dstName, err)
		}
	}
	// If the e2e test doesn't exist, L4 will simply have 0/0 results.

	return verifyDir, nil
}

// RunVerify executes verify.sh against a worktree directory.
// Returns the combined stdout+stderr output.
//
// Fix 2: Distinguishes test failures (ExitError, expected) from infrastructure
// errors (bash not found, permission denied, etc.) which are returned as errors.
func RunVerify(verifyDir string, worktreeDir string) (string, error) {
	verifyScript := filepath.Join(verifyDir, "verify.sh")

	cmd := exec.Command("bash", verifyScript, worktreeDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// ExitError means the script ran but exited non-zero (e.g. test failures).
		// That's expected; return the output for parsing.
		if _, ok := err.(*exec.ExitError); !ok {
			return string(out), fmt.Errorf("exec verify.sh: %w", err)
		}
	}
	return string(out), nil
}
