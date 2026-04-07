// Package adapter — v4-claude adapter implementation.
//
// Spec: .planning/workflow-bench-appendix.md (section A.3)
package adapter

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// V4ClaudeAdapter runs claude CLI with --agent manager for multi-agent workflows.
// It copies agent .md files and the reference/ directory into the worktree,
// creates the .planning/manager/ directory structure, then invokes claude with
// the manager agent to implement the plan.
//
// @implements REQ-ADAPTER-V4CLAUDE (v4-claude adapter: agent file setup + claude --agent manager)
type V4ClaudeAdapter struct {
	AgentsDir string // agent file source directory (default: ~/.claude/agents)
}

// NewV4Claude creates a new V4ClaudeAdapter from config.
// The config map may contain "agents_dir" to override the default ~/.claude/agents.
func NewV4Claude(cfg map[string]any) (Adapter, error) {
	agentsDir := defaultAgentsDir()
	if cfg != nil {
		if v, ok := cfg["agents_dir"]; ok {
			if s, ok := v.(string); ok && s != "" {
				agentsDir = expandHome(s)
			}
		}
	}
	return &V4ClaudeAdapter{AgentsDir: agentsDir}, nil
}

// Name returns the adapter name.
func (a *V4ClaudeAdapter) Name() string { return "v4-claude" }

// Setup prepares the worktree for the v4-claude workflow.
// It copies agent .md files to worktreeDir/.claude/agents/,
// copies the reference/ subdirectory (if it exists), and creates
// the .planning/manager/ directory structure that the manager agent expects.
//
// @implements REQ-ADAPTER-V4CLAUDE-SETUP (copy agent files + create planning directory)
func (a *V4ClaudeAdapter) Setup(ctx context.Context, worktreeDir string) error {
	agentsTarget := filepath.Join(worktreeDir, ".claude", "agents")
	if err := os.MkdirAll(agentsTarget, 0755); err != nil {
		return fmt.Errorf("create agents dir: %w", err)
	}

	entries, err := os.ReadDir(a.AgentsDir)
	if err != nil {
		return fmt.Errorf("read agents source dir %s: %w", a.AgentsDir, err)
	}

	for _, e := range entries {
		srcPath := filepath.Join(a.AgentsDir, e.Name())
		dstPath := filepath.Join(agentsTarget, e.Name())

		if e.IsDir() {
			if e.Name() == "reference" {
				if err := copyDir(srcPath, dstPath); err != nil {
					return fmt.Errorf("copy reference dir: %w", err)
				}
			}
			continue
		}
		if filepath.Ext(e.Name()) == ".md" {
			if err := copyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("copy agent file %s: %w", e.Name(), err)
			}
		}
	}

	// Create .planning/manager/ directory structure that manager agent expects.
	planningDir := filepath.Join(worktreeDir, ".planning", "manager")
	if err := os.MkdirAll(planningDir, 0755); err != nil {
		return fmt.Errorf("create planning dir: %w", err)
	}

	return nil
}

// Run executes the v4-claude workflow using claude --agent manager.
// The plan is written to a temporary file (same as vanilla) and passed
// via prompt to the manager agent. Output is parsed from JSON format.
//
// @implements REQ-ADAPTER-V4CLAUDE-RUN (plan file + claude --agent manager + JSON parse)
func (a *V4ClaudeAdapter) Run(ctx context.Context, worktreeDir string, planContent string) (*RunOutput, error) {
	start := time.Now()

	// Write plan to a temporary file to avoid ARG_MAX limits.
	planFile := filepath.Join(worktreeDir, ".bench-plan.md")
	if err := os.WriteFile(planFile, []byte(planContent), 0644); err != nil {
		return nil, fmt.Errorf("write plan file: %w", err)
	}
	defer os.Remove(planFile)

	planPrompt := fmt.Sprintf(
		"Read the approved plan from %s and implement it. This is a benchmark run — proceed without human approval gates.",
		planFile,
	)
	cmd := exec.CommandContext(ctx, "claude",
		"--agent", "manager",
		"-p", planPrompt,
		"--output-format", "json",
	)
	cmd.Dir = worktreeDir

	stdout, err := cmd.Output()
	wallTime := time.Since(start)

	result := &RunOutput{
		ExitCode: 0,
		Stdout:   string(stdout),
		WallTime: wallTime,
	}

	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
		result.Stderr = string(exitErr.Stderr)
	} else if err != nil {
		// Infrastructure error (e.g., claude not found).
		return nil, fmt.Errorf("exec claude: %w", err)
	}

	// Parse Claude JSON output for usage data.
	parseClaudeJSON(stdout, result)

	return result, nil
}

// defaultAgentsDir returns the default agents directory: ~/.claude/agents.
func defaultAgentsDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = os.TempDir()
	}
	return filepath.Join(home, ".claude", "agents")
}

// expandHome replaces a leading "~/" with the user's home directory.
func expandHome(path string) string {
	if len(path) < 2 || path[:2] != "~/" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = os.TempDir()
	}
	return filepath.Join(home, path[2:])
}

// copyFile copies a single file from src to dst, preserving permissions.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

// copyDir recursively copies a directory tree from src to dst.
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())

		if e.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
