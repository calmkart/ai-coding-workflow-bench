// Package adapter defines the interface for workflow adapters and a registry.
//
// Spec: .planning/workflow-bench.md, appendix A.1
package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// TokenUsage holds token consumption data from Claude CLI output.
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// RunOutput is the result of a single workflow execution.
type RunOutput struct {
	ExitCode   int
	Stdout     string
	Stderr     string
	WallTime   time.Duration
	TokenUsage *TokenUsage
	ToolUses   int
}

// Adapter defines the interface for workflow adapters.
//
// @implements REQ-ADAPTER-IFACE (adapter interface definition)
type Adapter interface {
	// Name returns a human-readable name for reports.
	Name() string

	// Setup prepares the worktree for this workflow (e.g., copy agent files).
	// Called once before Run. worktreeDir is the isolated git worktree path.
	Setup(ctx context.Context, worktreeDir string) error

	// Run executes the workflow against the given plan in the worktree.
	// planContent is the full text of the approved plan.md.
	// Returns RunOutput or error (error = infrastructure failure, not task failure).
	Run(ctx context.Context, worktreeDir string, planContent string) (*RunOutput, error)
}

// Registry maps adapter names to constructor functions.
var Registry = map[string]func(cfg map[string]any) (Adapter, error){
	"vanilla":   NewVanilla,
	"v4-claude": NewV4Claude,
	"custom":    NewCustom,
}

// Get returns an adapter by name from the registry.
func Get(name string, cfg map[string]any) (Adapter, error) {
	ctor, ok := Registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown adapter: %q", name)
	}
	return ctor(cfg)
}

// parseClaudeJSON extracts token usage and tool_uses from Claude CLI JSON output.
// If stdout is not valid JSON or doesn't contain the expected fields, result is
// left unchanged (no error — non-JSON output is valid for custom adapters).
func parseClaudeJSON(stdout []byte, result *RunOutput) {
	var parsed map[string]any
	if json.Unmarshal(stdout, &parsed) == nil {
		if usage, ok := parsed["usage"].(map[string]any); ok {
			result.TokenUsage = &TokenUsage{
				InputTokens:  jsonInt(usage["input_tokens"]),
				OutputTokens: jsonInt(usage["output_tokens"]),
			}
		}
		if tu, ok := parsed["tool_uses"].(float64); ok {
			result.ToolUses = int(tu)
		}
	}
}

// jsonInt safely converts a JSON number to int.
func jsonInt(v any) int {
	if f, ok := v.(float64); ok {
		return int(f)
	}
	return 0
}
