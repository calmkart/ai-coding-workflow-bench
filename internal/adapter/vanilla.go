// Package adapter — vanilla adapter implementation.
package adapter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// VanillaAdapter runs claude CLI directly with the plan.
type VanillaAdapter struct{}

// NewVanilla creates a new VanillaAdapter.
func NewVanilla(cfg map[string]any) (Adapter, error) {
	return &VanillaAdapter{}, nil
}

// Name returns the adapter name.
func (a *VanillaAdapter) Name() string { return "vanilla" }

// Setup is a no-op for vanilla adapter.
func (a *VanillaAdapter) Setup(ctx context.Context, worktreeDir string) error {
	return nil
}

// Run executes claude CLI with the plan content.
// The plan is written to a temporary file to avoid ARG_MAX limits.
// Claude CLI is called with --output-format json to capture token usage.
//
// @implements REQ-ADAPTER-VANILLA (vanilla adapter: plan file + claude -p + JSON parse)
func (a *VanillaAdapter) Run(ctx context.Context, worktreeDir string, planContent string) (*RunOutput, error) {
	start := time.Now()

	// Write plan to a temporary file to avoid ARG_MAX limits.
	planFile := filepath.Join(worktreeDir, ".bench-plan.md")
	if err := os.WriteFile(planFile, []byte(planContent), 0644); err != nil {
		return nil, fmt.Errorf("write plan file: %w", err)
	}
	defer os.Remove(planFile)

	planPrompt := fmt.Sprintf("Read the plan from %s and implement it.", planFile)
	cmd := exec.CommandContext(ctx, "claude",
		"-p", planPrompt,
		"--output-format", "json",
		"--dangerously-skip-permissions",
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
