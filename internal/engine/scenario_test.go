package engine

import (
	"strings"
	"testing"
)

// --- Collector: BENCH_RESULT in the middle of large output ---

func TestScenario_ParseVerifyOutput_ResultInMiddleOfLogs(t *testing.T) {
	var sb strings.Builder
	// Write 100 lines of noise before
	for i := 0; i < 100; i++ {
		sb.WriteString("=== some build output or test log line ===\n")
	}
	sb.WriteString("BENCH_RESULT: L1=PASS L2=6/8 L3=2 L4=3/5\n")
	// Write 50 lines of noise after
	for i := 0; i < 50; i++ {
		sb.WriteString("cleanup: removing temporary files...\n")
	}

	result, err := ParseVerifyOutput(sb.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.L1Build {
		t.Error("expected L1Build=true")
	}
	if result.L2Passed != 6 || result.L2Total != 8 {
		t.Errorf("L2: expected 6/8, got %d/%d", result.L2Passed, result.L2Total)
	}
	if result.L3Issues != 2 {
		t.Errorf("L3: expected 2, got %d", result.L3Issues)
	}
	if result.L4Passed != 3 || result.L4Total != 5 {
		t.Errorf("L4: expected 3/5, got %d/%d", result.L4Passed, result.L4Total)
	}
}

// --- Adversarial: Empty input ---

func TestScenario_ParseVerifyOutput_EmptyInput(t *testing.T) {
	_, err := ParseVerifyOutput("")
	if err == nil {
		t.Error("expected error for empty input")
	}
}

// --- Adversarial: Malformed BENCH_RESULT lines ---

func TestScenario_ParseVerifyOutput_MalformedL2(t *testing.T) {
	// L2 with non-numeric values should error because regex requires \d+
	input := "BENCH_RESULT: L1=PASS L2=abc/def L3=0 L4=0/0\n"
	_, err := ParseVerifyOutput(input)
	if err == nil {
		t.Error("expected error for malformed input")
	}
}

// --- Adversarial: Multiple BENCH_RESULT lines ---

func TestScenario_ParseVerifyOutput_MultipleBenchResults(t *testing.T) {
	// Multiple BENCH_RESULT lines with distinct L2 values.
	// FindStringSubmatch returns the first match.
	input := "BENCH_RESULT: L1=PASS L2=8/8 L3=0 L4=5/5\nBENCH_RESULT: L1=PASS L2=3/5 L3=2 L4=2/3\n"
	result, err := ParseVerifyOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.L1Build {
		t.Error("expected L1Build=true")
	}
	// Parser uses FindStringSubmatch which returns the first match.
	if result.L2Passed != 8 {
		t.Errorf("expected L2Passed=8 (first match), got %d", result.L2Passed)
	}
	if result.L2Total != 8 {
		t.Errorf("expected L2Total=8 (first match), got %d", result.L2Total)
	}
}

// --- Adversarial: High lint issue count ---

func TestScenario_ParseVerifyOutput_HighLintCount(t *testing.T) {
	input := "BENCH_RESULT: L1=PASS L2=10/10 L3=999 L4=5/5\n"
	result, err := ParseVerifyOutput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.L3Issues != 999 {
		t.Errorf("L3: expected 999, got %d", result.L3Issues)
	}
}
