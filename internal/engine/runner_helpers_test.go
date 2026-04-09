package engine

import (
	"bytes"
	"math"
	"os"
	"testing"
	"time"

	"github.com/calmkart/ai-coding-workflow-bench/internal/adapter"
	"github.com/calmkart/ai-coding-workflow-bench/internal/store"
)

// --- P16: Tests for extracted helper functions ---

// TestRecordTokens_WithTokenUsage verifies that recordTokens populates token
// fields, cost, and efficiency score on the Run when token data is available.
func TestRecordTokens_WithTokenUsage(t *testing.T) {
	run := &store.Run{}
	output := &adapter.RunOutput{
		WallTime: 30 * time.Second,
		ToolUses: 5,
		TokenUsage: &adapter.TokenUsage{
			InputTokens:  100000,
			OutputTokens: 50000,
		},
	}
	cfg := RunConfig{
		InputPricePerMTok:  3.0,
		OutputPricePerMTok: 15.0,
	}

	recordTokens(run, output, cfg, 1) // tier 1, budget $0.50

	// Token counts.
	if run.InputTokens == nil || *run.InputTokens != 100000 {
		t.Errorf("InputTokens: got %v, want 100000", run.InputTokens)
	}
	if run.OutputTokens == nil || *run.OutputTokens != 50000 {
		t.Errorf("OutputTokens: got %v, want 50000", run.OutputTokens)
	}
	if run.TotalTokens == nil || *run.TotalTokens != 150000 {
		t.Errorf("TotalTokens: got %v, want 150000", run.TotalTokens)
	}

	// Cost: (100000/1M)*3.0 + (50000/1M)*15.0 = 0.3 + 0.75 = 1.05
	if run.CostUSD == nil || math.Abs(*run.CostUSD-1.05) > 0.001 {
		t.Errorf("CostUSD: got %v, want ~1.05", run.CostUSD)
	}

	// Efficiency: 1.0 - min(1.0, 1.05/0.50) = 1.0 - 1.0 = 0.0 (over budget)
	if run.EfficiencyScore == nil || math.Abs(*run.EfficiencyScore-0.0) > 0.001 {
		t.Errorf("EfficiencyScore: got %v, want 0.0", run.EfficiencyScore)
	}

	// Wall time.
	if run.WallTimeSecs == nil || math.Abs(*run.WallTimeSecs-30.0) > 0.001 {
		t.Errorf("WallTimeSecs: got %v, want 30.0", run.WallTimeSecs)
	}

	// Tool uses.
	if run.ToolUses == nil || *run.ToolUses != 5 {
		t.Errorf("ToolUses: got %v, want 5", run.ToolUses)
	}
}

// TestRecordTokens_NoTokenUsage verifies that recordTokens handles nil TokenUsage
// gracefully, only setting wall time and tool uses.
func TestRecordTokens_NoTokenUsage(t *testing.T) {
	run := &store.Run{}
	output := &adapter.RunOutput{
		WallTime:   10 * time.Second,
		ToolUses:   0,
		TokenUsage: nil,
	}
	cfg := RunConfig{}

	recordTokens(run, output, cfg, 1)

	// Token fields should remain nil.
	if run.InputTokens != nil {
		t.Error("InputTokens should be nil when no token usage")
	}
	if run.CostUSD != nil {
		t.Error("CostUSD should be nil when no token usage")
	}
	if run.EfficiencyScore != nil {
		t.Error("EfficiencyScore should be nil when no token usage")
	}

	// Wall time should still be set.
	if run.WallTimeSecs == nil || math.Abs(*run.WallTimeSecs-10.0) > 0.001 {
		t.Errorf("WallTimeSecs: got %v, want 10.0", run.WallTimeSecs)
	}

	// ToolUses should be nil when 0.
	if run.ToolUses != nil {
		t.Error("ToolUses should be nil when output.ToolUses is 0")
	}
}

// TestRecordTokens_UnderBudget verifies efficiency score for a run under budget.
func TestRecordTokens_UnderBudget(t *testing.T) {
	run := &store.Run{}
	output := &adapter.RunOutput{
		WallTime: 5 * time.Second,
		TokenUsage: &adapter.TokenUsage{
			InputTokens:  10000,
			OutputTokens: 5000,
		},
	}
	cfg := RunConfig{
		InputPricePerMTok:  3.0,
		OutputPricePerMTok: 15.0,
	}

	recordTokens(run, output, cfg, 2) // tier 2, budget $1.00

	// Cost: (10000/1M)*3.0 + (5000/1M)*15.0 = 0.03 + 0.075 = 0.105
	// Efficiency: 1.0 - min(1.0, 0.105/1.00) = 1.0 - 0.105 = 0.895
	if run.EfficiencyScore == nil || math.Abs(*run.EfficiencyScore-0.895) > 0.001 {
		t.Errorf("EfficiencyScore: got %v, want ~0.895", run.EfficiencyScore)
	}
}

// TestPrintRunResult_DoesNotPanic verifies that printRunResult doesn't panic
// for various inputs.
func TestPrintRunResult_DoesNotPanic(t *testing.T) {
	// Redirect stdout to avoid cluttering test output.
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	defer func() {
		w.Close()
		os.Stdout = old
	}()

	run := &store.Run{Status: "completed"}
	vr := &VerifyResult{
		L1Build:  true,
		L2Passed: 8, L2Total: 8,
		L3Issues: 0,
		L4Passed: 5, L4Total: 5,
	}
	output := &adapter.RunOutput{
		WallTime: 30 * time.Second,
		ExitCode: 0,
	}

	// Should not panic, no mutex.
	printRunResult(run, vr, output, 0.95, nil)

	// With RubricComposite set.
	v := 4.2
	run.RubricComposite = &v
	printRunResult(run, vr, output, 0.95, nil)
}

// TestPrintRunResult_L1Fail verifies that printRunResult shows L1=FAIL.
func TestPrintRunResult_L1Fail(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	run := &store.Run{Status: "completed"}
	vr := &VerifyResult{
		L1Build:  false,
		L2Passed: 0, L2Total: 0,
		L4Passed: 0, L4Total: 0,
	}
	output := &adapter.RunOutput{
		WallTime: 5 * time.Second,
	}

	printRunResult(run, vr, output, 0.0, nil)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if !bytes.Contains([]byte(out), []byte("L1=FAIL")) {
		t.Errorf("expected L1=FAIL in output, got:\n%s", out)
	}
}
