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

//go:embed templates/http_server.sh.tmpl
var httpServerTemplate string

// VerifyConfig holds parameters for generating a verify script.
type VerifyConfig struct {
	TaskType string // http-server, k8s-operator, library, cli
	TaskDir  string // path to task directory (contains verify/e2e_test.go)
	RunID    string
}

// GenerateVerifyDir creates a temporary directory with verify.sh and e2e_test.go.
// Returns the path to the temporary directory.
//
// @implements REQ-VERIFY-GEN (generate verify.sh from template + copy e2e_test.go)
func GenerateVerifyDir(cfg VerifyConfig) (string, error) {
	verifyDir := filepath.Join(os.TempDir(), "bench-verify-"+cfg.RunID)
	if err := os.MkdirAll(verifyDir, 0755); err != nil {
		return "", fmt.Errorf("create verify dir: %w", err)
	}

	// Generate verify.sh from template.
	var tmplStr string
	switch cfg.TaskType {
	case "http-server":
		tmplStr = httpServerTemplate
	default:
		// All task types use the same generic verify template (build + test + vet + e2e)
		slog.Info("using generic verify template", "type", cfg.TaskType)
		tmplStr = httpServerTemplate
	}
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

	// Copy e2e_test.go from task's verify/ directory.
	// Source file uses .src extension to avoid Go tooling trying to compile it in place.
	e2eSrc := filepath.Join(cfg.TaskDir, "verify", "e2e_test.go.src")
	e2eDst := filepath.Join(verifyDir, "e2e_test.go")
	if data, err := os.ReadFile(e2eSrc); err == nil {
		if err := os.WriteFile(e2eDst, data, 0644); err != nil {
			return "", fmt.Errorf("copy e2e_test.go: %w", err)
		}
	}
	// If e2e_test.go doesn't exist, L4 will simply have 0/0 results.

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
