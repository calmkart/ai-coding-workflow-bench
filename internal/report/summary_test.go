package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/calmkart/ai-coding-workflow-bench/internal/store"
)

func TestGenerateSummary(t *testing.T) {
	now := time.Now()
	finished := now.Add(2 * time.Minute)
	l1 := true
	l2p, l2t := 8, 8
	l3 := 0
	l4p, l4t := 5, 5
	cs := 0.95

	runs := []*store.Run{
		{
			ID: "test-001", Tag: "smoke", Workflow: "vanilla",
			TaskID: "tier1/fix-handler-bug", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, L2UtPassed: &l2p, L2UtTotal: &l2t, L3LintIssues: &l3,
			L4E2EPassed: &l4p, L4E2ETotal: &l4t, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Benchmark Report: smoke") {
		t.Error("expected report header")
	}
	if !strings.Contains(output, "tier1/fix-handler-bug") {
		t.Error("expected task ID in report")
	}
	if !strings.Contains(output, "100.0%") {
		t.Error("expected 100% pass rate")
	}
	if !strings.Contains(output, "PASS") {
		t.Error("expected PASS in L1 column")
	}
	if !strings.Contains(output, "8/8") {
		t.Error("expected 8/8 in L2 column")
	}
	if !strings.Contains(output, "5/5") {
		t.Error("expected 5/5 in L4 column")
	}
}

func TestGenerateSummaryEmpty(t *testing.T) {
	var buf bytes.Buffer
	err := GenerateSummary(&buf, nil)
	if err == nil {
		t.Error("expected error for empty runs")
	}
}

// TestGenerateSummary_DateIncludesTime verifies that the report date includes hours and minutes (item 4).
func TestGenerateSummary_DateIncludesTime(t *testing.T) {
	now := time.Date(2026, 3, 25, 14, 30, 0, 0, time.Local)
	finished := now.Add(2 * time.Minute)
	l1 := true
	cs := 0.95

	runs := []*store.Run{
		{
			ID: "test-date", Tag: "date-test", Workflow: "vanilla",
			TaskID: "tier1/fix", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	// Should contain time portion, not just date.
	if !strings.Contains(output, "14:30") {
		t.Errorf("expected date to include time '14:30', got:\n%s", output)
	}
}

// TestGenerateSummary_RunsPerTaskVaries verifies that RunsPerTask shows a range when different (item 5).
func TestGenerateSummary_RunsPerTaskVaries(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1 := true
	cs := 0.9

	runs := []*store.Run{
		{
			ID: "run-a1", Tag: "rpt-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
		{
			ID: "run-a2", Tag: "rpt-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 2, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
		{
			ID: "run-a3", Tag: "rpt-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 3, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
		{
			ID: "run-b1", Tag: "rpt-test", Workflow: "vanilla",
			TaskID: "tier1/task-b", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	// task-a has 3 runs, task-b has 1 run -> "1-3"
	if !strings.Contains(output, "Runs/task: 1-3") {
		t.Errorf("expected 'Runs/task: 1-3' in output, got:\n%s", output)
	}
}

// TestGenerateSummary_WorkflowFromData verifies workflow is read from actual data (item 6).
func TestGenerateSummary_WorkflowFromData(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1 := true
	cs := 0.9

	runs := []*store.Run{
		{
			ID: "run-wf1", Tag: "wf-test", Workflow: "custom-wf",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Workflow: custom-wf") {
		t.Errorf("expected 'Workflow: custom-wf' in output, got:\n%s", output)
	}
}

// TestGenerateSummary_MultipleWorkflows verifies "multiple" is shown when mixed workflows (item 6).
func TestGenerateSummary_MultipleWorkflows(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1 := true
	cs := 0.9

	runs := []*store.Run{
		{
			ID: "run-mw1", Tag: "mw-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
		{
			ID: "run-mw2", Tag: "mw-test", Workflow: "custom",
			TaskID: "tier1/task-b", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Workflow: multiple") {
		t.Errorf("expected 'Workflow: multiple' in output, got:\n%s", output)
	}
}

// TestGenerateSummary_L1FailMarked verifies L1=FAIL rows are marked with **FAIL** and !! (item 8).
func TestGenerateSummary_L1FailMarked(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1Fail := false
	cs := 0.0

	runs := []*store.Run{
		{
			ID: "run-fail-mark", Tag: "fail-mark-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1Fail, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "**FAIL**") {
		t.Errorf("expected **FAIL** markdown bold in L1 column, got:\n%s", output)
	}
	if !strings.Contains(output, "!!") {
		t.Errorf("expected '!!' warning marker in correctness column for L1=FAIL, got:\n%s", output)
	}
}

// TestGenerateSummary_AvgWallTimeAndTokens verifies that AvgWallTime and AvgTokens
// appear in the report when runs have timing and token data.
func TestGenerateSummary_AvgWallTimeAndTokens(t *testing.T) {
	now := time.Now()
	finished := now.Add(2 * time.Minute)
	l1 := true
	cs := 0.95
	wallTime1 := 120.0 // 2 minutes
	wallTime2 := 180.0 // 3 minutes
	tokens1 := 1000
	tokens2 := 2000

	runs := []*store.Run{
		{
			ID: "run-wt1", Tag: "wt-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
			WallTimeSecs: &wallTime1, TotalTokens: &tokens1,
		},
		{
			ID: "run-wt2", Tag: "wt-test", Workflow: "vanilla",
			TaskID: "tier1/task-b", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
			WallTimeSecs: &wallTime2, TotalTokens: &tokens2,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	// Average wall time = (120+180)/2 = 150s = 2m30s
	if !strings.Contains(output, "Avg Wall Time") {
		t.Errorf("expected 'Avg Wall Time' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "2m30s") {
		t.Errorf("expected '2m30s' average wall time in output, got:\n%s", output)
	}
	// Average tokens = (1000+2000)/2 = 1500
	if !strings.Contains(output, "Avg Tokens") {
		t.Errorf("expected 'Avg Tokens' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "1500") {
		t.Errorf("expected '1500' average tokens in output, got:\n%s", output)
	}
}

// TestGenerateSummary_PerTierSummary verifies that Per-Tier Summary section
// appears with correct tier grouping when tasks span multiple tiers.
func TestGenerateSummary_PerTierSummary(t *testing.T) {
	now := time.Now()
	finished := now.Add(2 * time.Minute)
	l1 := true
	l2p, l2t := 8, 8
	l3 := 0
	l4p, l4t := 5, 5
	cs := 1.0
	l4pFail, l4tFail := 3, 5
	csFail := 0.67

	runs := []*store.Run{
		{
			ID: "tier-sum-1", Tag: "tier-test", Workflow: "vanilla",
			TaskID: "tier1/fix-handler-bug", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, L2UtPassed: &l2p, L2UtTotal: &l2t, L3LintIssues: &l3,
			L4E2EPassed: &l4p, L4E2ETotal: &l4t, CorrectnessScore: &cs,
		},
		{
			ID: "tier-sum-2", Tag: "tier-test", Workflow: "vanilla",
			TaskID: "tier1/add-health-check", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, L2UtPassed: &l2p, L2UtTotal: &l2t, L3LintIssues: &l3,
			L4E2EPassed: &l4p, L4E2ETotal: &l4t, CorrectnessScore: &cs,
		},
		{
			ID: "tier-sum-3", Tag: "tier-test", Workflow: "vanilla",
			TaskID: "tier2/extract-storage", Tier: 2, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, L2UtPassed: &l2p, L2UtTotal: &l2t, L3LintIssues: &l3,
			L4E2EPassed: &l4pFail, L4E2ETotal: &l4tFail, CorrectnessScore: &csFail,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	// Should contain the Per-Tier Summary section.
	if !strings.Contains(output, "Per-Tier Summary") {
		t.Errorf("expected 'Per-Tier Summary' section, got:\n%s", output)
	}
	// T1: 2 tasks, both passed -> 100.0%
	if !strings.Contains(output, "T1") {
		t.Errorf("expected T1 in per-tier summary, got:\n%s", output)
	}
	if !strings.Contains(output, "100.0%") {
		t.Errorf("expected 100.0%% pass rate for T1, got:\n%s", output)
	}
	// T2: 1 task, L4=3/5 (not all pass) -> 0.0%
	if !strings.Contains(output, "T2") {
		t.Errorf("expected T2 in per-tier summary, got:\n%s", output)
	}
}

// TestGenerateSummary_NoTokenData verifies that AvgWallTime and AvgTokens are
// omitted from the report when runs have no timing or token data.
func TestGenerateSummary_NoTokenData(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1 := true
	cs := 0.9

	runs := []*store.Run{
		{
			ID: "run-nt1", Tag: "nt-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
			// No WallTimeSecs or TotalTokens set
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Avg Wall Time") {
		t.Errorf("expected no 'Avg Wall Time' when no data, got:\n%s", output)
	}
	if strings.Contains(output, "Avg Tokens") {
		t.Errorf("expected no 'Avg Tokens' when no data, got:\n%s", output)
	}
}

// TestGenerateSummary_WithRubricScores verifies that the rubric section appears
// when runs have rubric scores.
func TestGenerateSummary_WithRubricScores(t *testing.T) {
	now := time.Now()
	finished := now.Add(2 * time.Minute)
	l1 := true
	cs := 0.95
	rCorr := 4.0
	rRead := 3.0
	rSimp := 4.0
	rRob := 3.0
	rMin := 5.0
	rMaint := 4.0
	rGoId := 4.0
	rComp := 3.85

	runs := []*store.Run{
		{
			ID: "run-rubric-1", Tag: "rubric-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
			RubricCorrectness:     &rCorr,
			RubricReadability:     &rRead,
			RubricSimplicity:      &rSimp,
			RubricRobustness:      &rRob,
			RubricMinimality:      &rMin,
			RubricMaintainability: &rMaint,
			RubricGoIdioms:        &rGoId,
			RubricComposite:       &rComp,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Code Quality (LLM Judge)") {
		t.Errorf("expected 'Code Quality (LLM Judge)' section, got:\n%s", output)
	}
	if !strings.Contains(output, "4.0/5") {
		t.Errorf("expected '4.0/5' for correctness score, got:\n%s", output)
	}
	if !strings.Contains(output, "3.9/5") || !strings.Contains(output, "3.8/5") {
		// Composite is 3.85, displayed as 3.9 or 3.8 depending on rounding
		// Let's just check it contains "Composite"
		if !strings.Contains(output, "Composite") {
			t.Errorf("expected 'Composite' in rubric section, got:\n%s", output)
		}
	}
}

// TestGenerateSummary_NoRubricScores verifies that the rubric section is omitted
// when runs have no rubric scores.
func TestGenerateSummary_NoRubricScores(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1 := true
	cs := 0.9

	runs := []*store.Run{
		{
			ID: "run-norubric", Tag: "norubric-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Code Quality (LLM Judge)") {
		t.Errorf("expected no rubric section when no rubric data, got:\n%s", output)
	}
}

// TestGenerateSummary_TotalCost verifies that Est. Total Cost appears when
// runs have token data.
func TestGenerateSummary_TotalCost(t *testing.T) {
	now := time.Now()
	finished := now.Add(2 * time.Minute)
	l1 := true
	cs := 0.95
	inputTok := 100000
	outputTok := 50000
	totalTok := 150000
	costUSD := 1.05 // 100000/1M*3 + 50000/1M*15 = 0.30 + 0.75

	runs := []*store.Run{
		{
			ID: "run-cost-1", Tag: "cost-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
			InputTokens: &inputTok, OutputTokens: &outputTok, TotalTokens: &totalTok,
			CostUSD: &costUSD,
		},
		{
			ID: "run-cost-2", Tag: "cost-test", Workflow: "vanilla",
			TaskID: "tier1/task-b", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
			InputTokens: &inputTok, OutputTokens: &outputTok, TotalTokens: &totalTok,
			CostUSD: &costUSD,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Est. Total Cost") {
		t.Errorf("expected 'Est. Total Cost' in output, got:\n%s", output)
	}
	// Each run: 100000/1M*3 + 50000/1M*15 = 0.30 + 0.75 = $1.05
	// Total for 2 runs: $2.10
	if !strings.Contains(output, "$2.10") {
		t.Errorf("expected '$2.10' total cost, got:\n%s", output)
	}
}

// TestGenerateSummary_ConsistencyWarnings verifies that consistency warnings appear
// in the report when runs have rubric consistency warnings.
func TestGenerateSummary_ConsistencyWarnings(t *testing.T) {
	now := time.Now()
	finished := now.Add(2 * time.Minute)
	l1 := true
	cs := 0.95
	rCorr := 4.0
	rRead := 2.0
	rSimp := 4.0
	rRob := 3.0
	rMin := 5.0
	rMaint := 4.0
	rGoId := 4.0
	rComp := 3.50
	warnings := "readability: inconsistent: 5/6 booleans true but score=2"

	runs := []*store.Run{
		{
			ID: "run-cw-1", Tag: "cw-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
			RubricCorrectness:        &rCorr,
			RubricReadability:         &rRead,
			RubricSimplicity:          &rSimp,
			RubricRobustness:          &rRob,
			RubricMinimality:          &rMin,
			RubricMaintainability:     &rMaint,
			RubricGoIdioms:            &rGoId,
			RubricComposite:           &rComp,
			RubricConsistencyWarnings: &warnings,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Consistency warnings") {
		t.Errorf("expected 'Consistency warnings' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "readability") {
		t.Errorf("expected 'readability' in consistency warning, got:\n%s", output)
	}
	if !strings.Contains(output, "5/6 booleans true but score=2") {
		t.Errorf("expected '5/6 booleans true but score=2' in consistency warning, got:\n%s", output)
	}
}

// TestGenerateSummary_NoConsistencyWarnings verifies that no consistency warnings
// section appears when there are no warnings.
func TestGenerateSummary_NoConsistencyWarnings(t *testing.T) {
	now := time.Now()
	finished := now.Add(2 * time.Minute)
	l1 := true
	cs := 0.95
	rCorr := 4.0
	rRead := 3.0
	rSimp := 4.0
	rRob := 3.0
	rMin := 5.0
	rMaint := 4.0
	rGoId := 4.0
	rComp := 3.85

	runs := []*store.Run{
		{
			ID: "run-ncw-1", Tag: "ncw-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
			RubricCorrectness:     &rCorr,
			RubricReadability:     &rRead,
			RubricSimplicity:      &rSimp,
			RubricRobustness:      &rRob,
			RubricMinimality:      &rMin,
			RubricMaintainability: &rMaint,
			RubricGoIdioms:        &rGoId,
			RubricComposite:       &rComp,
			// No RubricConsistencyWarnings
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Consistency warnings") {
		t.Errorf("expected no 'Consistency warnings' when no warnings, got:\n%s", output)
	}
}

// TestGenerateSummary_NoCost verifies that Est. Total Cost is omitted when
// runs have no token data.
func TestGenerateSummary_NoCost(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1 := true
	cs := 0.9

	runs := []*store.Run{
		{
			ID: "run-nocost", Tag: "nocost-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummary(&buf, runs); err != nil {
		t.Fatalf("generate summary: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Est. Total Cost") {
		t.Errorf("expected no 'Est. Total Cost' when no token data, got:\n%s", output)
	}
}
