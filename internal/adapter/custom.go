// Package adapter — custom adapter implementation.
//
// Spec: .planning/workflow-bench-appendix.md (section A.4)
package adapter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// CustomAdapter executes a user-defined command as the workflow entry point.
// It allows users to configure arbitrary commands in bench.yaml without writing Go code.
// Setup commands are run in order before the entry command, all within the worktree directory.
// The entry command receives context via environment variables (BENCH_REPO_DIR, BENCH_PLAN_FILE,
// BENCH_PLAN_PROMPT) and its stdout is parsed for optional JSON token usage data.
//
// @implements REQ-ADAPTER-CUSTOM (custom adapter: user-defined command execution via bench.yaml)
type CustomAdapter struct {
	name         string   // workflow name
	entryCommand  string   // main execution command (run via bash -c)
	setupCommands []string // preparation commands (run in worktree before entry_command)
}

// NewCustom creates a new CustomAdapter from config.
// The config map must contain "entry_command" (required) and may contain
// "setup_commands" (optional list of strings) and "name" (optional).
//
// @implements REQ-ADAPTER-CUSTOM-CTOR (custom adapter constructor: validate entry_command)
func NewCustom(cfg map[string]any) (Adapter, error) {
	entryCmd := ""
	if cfg != nil {
		if v, ok := cfg["entry_command"]; ok {
			if s, ok := v.(string); ok {
				entryCmd = s
			}
		}
	}
	if entryCmd == "" {
		return nil, fmt.Errorf("custom adapter requires entry_command")
	}

	name := "custom"
	if cfg != nil {
		if v, ok := cfg["name"]; ok {
			if s, ok := v.(string); ok && s != "" {
				name = s
			}
		}
	}

	var setupCmds []string
	if cfg != nil {
		if v, ok := cfg["setup_commands"]; ok {
			switch cmds := v.(type) {
			case []string:
				setupCmds = cmds
			case []any:
				for _, item := range cmds {
					if s, ok := item.(string); ok {
						setupCmds = append(setupCmds, s)
					}
				}
			}
		}
	}

	return &CustomAdapter{
		name:         name,
		entryCommand:  entryCmd,
		setupCommands: setupCmds,
	}, nil
}

// Name returns the adapter name.
func (a *CustomAdapter) Name() string { return a.name }

// Setup executes setup_commands sequentially in the worktree directory.
// Each command is run via "bash -c" so users can use pipes, redirects, etc.
// If any command fails, Setup returns an error containing the command and its output.
//
// @implements REQ-ADAPTER-CUSTOM-SETUP (execute setup_commands in worktree before entry_command)
func (a *CustomAdapter) Setup(ctx context.Context, worktreeDir string) error {
	for _, setupCmd := range a.setupCommands {
		cmd := exec.CommandContext(ctx, "bash", "-c", setupCmd)
		cmd.Dir = worktreeDir
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("setup command %q failed: %w\noutput: %s", setupCmd, err, string(out))
		}
	}
	return nil
}

// Run executes the user-defined entry_command in the worktree directory.
// The plan is written to .bench-plan.md, and context is passed via environment variables:
//   - BENCH_REPO_DIR: absolute path to the worktree
//   - BENCH_PLAN_FILE: absolute path to the plan file
//   - BENCH_PLAN_PROMPT: convenience prompt string for use in entry_command
//
// If stdout contains valid JSON with a "usage" field, token data is extracted.
// An ExitError (non-zero exit) is treated as a task failure (not infrastructure error).
// Other errors (e.g., bash not found) are treated as infrastructure errors.
//
// @implements REQ-ADAPTER-CUSTOM-RUN (plan file + env vars + bash -c entry_command + JSON parse)
func (a *CustomAdapter) Run(ctx context.Context, worktreeDir string, planContent string) (*RunOutput, error) {
	start := time.Now()

	// Write plan to a temporary file to avoid ARG_MAX limits.
	planFile := filepath.Join(worktreeDir, ".bench-plan.md")
	if err := os.WriteFile(planFile, []byte(planContent), 0644); err != nil {
		return nil, fmt.Errorf("write plan file: %w", err)
	}
	defer os.Remove(planFile)

	planPrompt := fmt.Sprintf("Read the plan from %s and implement it.", planFile)

	cmd := exec.CommandContext(ctx, "bash", "-c", a.entryCommand)
	cmd.Dir = worktreeDir
	cmd.Env = append(os.Environ(),
		"BENCH_REPO_DIR="+worktreeDir,
		"BENCH_PLAN_FILE="+planFile,
		"BENCH_PLAN_PROMPT="+planPrompt,
	)

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
		// Infrastructure error (e.g., bash not found).
		return nil, fmt.Errorf("exec custom command: %w", err)
	}

	// Parse stdout for optional JSON token usage data.
	parseClaudeJSON(stdout, result)

	return result, nil
}
