package metrics

import (
	"testing"
)

// --- Partial Pass ---
// L2=5/10 (0.5), L3=2 issues (1.0-2*0.05=0.9), L4=3/5 (0.6)
// correctness = 0.20*0.5 + 0.10*0.9 + 0.70*0.6 = 0.10 + 0.09 + 0.42 = 0.61

func TestScenario_Correctness_PartialPass(t *testing.T) {
	got := CalculateCorrectness(CorrectnessInput{
		L1Build:  true,
		L2Passed: 5, L2Total: 10,
		L3Issues: 2,
		L4Passed: 3, L4Total: 5,
	})
	if !approxEqual(got, 0.61) {
		t.Errorf("partial pass: expected 0.61, got %.4f", got)
	}
}

// --- Adversarial: L1 FAIL with zero everything ---

func TestScenario_Correctness_L1FailZeroEverything(t *testing.T) {
	got := CalculateCorrectness(CorrectnessInput{
		L1Build:  false,
		L2Passed: 0, L2Total: 0,
		L3Issues: 0,
		L4Passed: 0, L4Total: 0,
	})
	if got != 0.0 {
		t.Errorf("L1 fail zero everything: expected 0.0, got %.4f", got)
	}
}

// --- Adversarial: Extreme values ---

func TestScenario_Correctness_ExtremelyHighLintIssues(t *testing.T) {
	got := CalculateCorrectness(CorrectnessInput{
		L1Build:  true,
		L2Passed: 100, L2Total: 100,
		L3Issues: 10000,
		L4Passed: 100, L4Total: 100,
	})
	// L3 score = max(0, 1.0 - 10000*0.05) = 0.0
	// 0.20*1.0 + 0.10*0.0 + 0.70*1.0 = 0.90
	if !approxEqual(got, 0.90) {
		t.Errorf("extreme lint issues: expected 0.90, got %.4f", got)
	}
}

// --- Adversarial: VT count exceeds base score massively ---

func TestScenario_Correctness_MassiveVTDeduction(t *testing.T) {
	got := CalculateCorrectness(CorrectnessInput{
		L1Build:  true,
		L2Passed: 8, L2Total: 8,
		L3Issues: 0,
		L4Passed: 5, L4Total: 5,
		CriticalVTFailCount: 100,
	})
	// 1.0 - 10.0 = -9.0, clamped to 0.0
	if got != 0.0 {
		t.Errorf("massive VT deduction: expected 0.0, got %.4f", got)
	}
}

// --- Bug fix: E2E compile failure should not yield perfect score ---
// When e2e_test.go exists but compilation fails, the verify template now
// produces L4=0/1 instead of L4=0/0. This test confirms correctness < 1.0.

func TestScenario_Correctness_E2ECompileFailure(t *testing.T) {
	got := CalculateCorrectness(CorrectnessInput{
		L1Build:  true,
		L2Passed: 8, L2Total: 8,
		L3Issues: 0,
		L4Passed: 0, L4Total: 1,
	})
	// L4=0/1 -> l4Score = 0.0
	// 0.20*1.0 + 0.10*1.0 + 0.70*0.0 = 0.30
	if !approxEqual(got, 0.30) {
		t.Errorf("E2E compile failure: expected 0.30, got %.4f", got)
	}
	// Critical: must be less than 1.0 (the old bug gave 1.0 for L4=0/0)
	if got >= 1.0 {
		t.Errorf("E2E compile failure must NOT yield perfect score, got %.4f", got)
	}
}

// --- Boundary: Exactly at lint threshold ---

func TestScenario_Correctness_ExactlyTwentyLintIssues(t *testing.T) {
	got := CalculateCorrectness(CorrectnessInput{
		L1Build:  true,
		L2Passed: 10, L2Total: 10,
		L3Issues: 20,
		L4Passed: 5, L4Total: 5,
	})
	// L3 = max(0, 1.0 - 20*0.05) = max(0, 0.0) = 0.0
	// 0.20*1.0 + 0.10*0.0 + 0.70*1.0 = 0.90
	if !approxEqual(got, 0.90) {
		t.Errorf("exactly 20 lint issues: expected 0.90, got %.4f", got)
	}
}
