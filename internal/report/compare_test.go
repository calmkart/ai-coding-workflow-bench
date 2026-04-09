package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/calmkart/ai-coding-workflow-bench/internal/metrics"
	"github.com/calmkart/ai-coding-workflow-bench/internal/store"
)

// helper to create a pointer to a value.
func ptrBool(b bool) *bool       { return &b }
func ptrInt(i int) *int          { return &i }
func ptrFloat(f float64) *float64 { return &f }

func makeRun(id, tag, taskID string, l1 bool, l2p, l2t, l4p, l4t int, corr float64, wallSecs float64) *store.Run {
	now := time.Now()
	finished := now.Add(time.Duration(wallSecs) * time.Second)
	r := &store.Run{
		ID:               id,
		Tag:              tag,
		Workflow:         "vanilla",
		TaskID:           taskID,
		Tier:             1,
		TaskType:         "http-server",
		RunNumber:        1,
		Status:           "completed",
		StartedAt:        now,
		FinishedAt:       &finished,
		L1Build:          ptrBool(l1),
		L2UtPassed:       ptrInt(l2p),
		L2UtTotal:        ptrInt(l2t),
		L4E2EPassed:      ptrInt(l4p),
		L4E2ETotal:       ptrInt(l4t),
		CorrectnessScore: ptrFloat(corr),
		WallTimeSecs:     ptrFloat(wallSecs),
	}
	return r
}

func TestGenerateComparison_Basic(t *testing.T) {
	leftRuns := []*store.Run{
		makeRun("l1", "vanilla-tag", "tier1/fix-handler-bug", true, 8, 8, 5, 5, 1.0, 64),
		makeRun("l2", "vanilla-tag", "tier1/fix-status-code", true, 8, 8, 5, 5, 1.0, 63),
	}
	rightRuns := []*store.Run{
		makeRun("r1", "v4-tag", "tier1/fix-handler-bug", false, 2, 8, 0, 5, 0.30, 912),
		makeRun("r2", "v4-tag", "tier1/fix-status-code", true, 8, 8, 5, 5, 1.0, 750),
	}

	var buf bytes.Buffer
	err := GenerateComparison(&buf, leftRuns, rightRuns, "vanilla-tag", "v4-tag")
	if err != nil {
		t.Fatalf("GenerateComparison: %v", err)
	}

	output := buf.String()

	// Check header.
	if !strings.Contains(output, "Comparison: vanilla-tag vs v4-tag") {
		t.Error("expected comparison header")
	}

	// Check both tags appear in the Overall table.
	if !strings.Contains(output, "vanilla-tag") {
		t.Error("expected left tag in output")
	}
	if !strings.Contains(output, "v4-tag") {
		t.Error("expected right tag in output")
	}

	// Check pass rates: left = 100%, right = 50%.
	if !strings.Contains(output, "100.0%") {
		t.Error("expected 100.0% pass rate for left")
	}
	if !strings.Contains(output, "50.0%") {
		t.Error("expected 50.0% pass rate for right")
	}

	// Check per-task section.
	if !strings.Contains(output, "tier1/fix-handler-bug") {
		t.Error("expected fix-handler-bug task in output")
	}
	if !strings.Contains(output, "tier1/fix-status-code") {
		t.Error("expected fix-status-code task in output")
	}

	// Check winner: fix-handler-bug should be "Left" (1.0 > 0.3).
	if !strings.Contains(output, "Left") {
		t.Error("expected 'Left' winner for fix-handler-bug")
	}

	// Check summary line.
	if !strings.Contains(output, "Left wins: 1") {
		t.Errorf("expected 'Left wins: 1', got:\n%s", output)
	}
	if !strings.Contains(output, "Tie: 1") {
		t.Errorf("expected 'Tie: 1', got:\n%s", output)
	}
}

func TestGenerateComparison_EmptyLeft(t *testing.T) {
	rightRuns := []*store.Run{
		makeRun("r1", "v4-tag", "tier1/fix", true, 8, 8, 5, 5, 1.0, 60),
	}

	var buf bytes.Buffer
	err := GenerateComparison(&buf, nil, rightRuns, "empty-tag", "v4-tag")
	if err == nil {
		t.Fatal("expected error for empty left runs")
	}
	if !strings.Contains(err.Error(), "no runs found for left tag") {
		t.Errorf("expected 'no runs found for left tag' error, got: %v", err)
	}
}

func TestGenerateComparison_EmptyRight(t *testing.T) {
	leftRuns := []*store.Run{
		makeRun("l1", "vanilla-tag", "tier1/fix", true, 8, 8, 5, 5, 1.0, 60),
	}

	var buf bytes.Buffer
	err := GenerateComparison(&buf, leftRuns, nil, "vanilla-tag", "empty-tag")
	if err == nil {
		t.Fatal("expected error for empty right runs")
	}
	if !strings.Contains(err.Error(), "no runs found for right tag") {
		t.Errorf("expected 'no runs found for right tag' error, got: %v", err)
	}
}

func TestGenerateComparison_AllTie(t *testing.T) {
	leftRuns := []*store.Run{
		makeRun("l1", "left", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 60),
		makeRun("l2", "left", "tier1/task-b", true, 8, 8, 5, 5, 1.0, 70),
	}
	rightRuns := []*store.Run{
		makeRun("r1", "right", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 90),
		makeRun("r2", "right", "tier1/task-b", true, 8, 8, 5, 5, 1.0, 80),
	}

	var buf bytes.Buffer
	err := GenerateComparison(&buf, leftRuns, rightRuns, "left", "right")
	if err != nil {
		t.Fatalf("GenerateComparison: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Tie: 2") {
		t.Errorf("expected 'Tie: 2', got:\n%s", output)
	}
	if !strings.Contains(output, "Left wins: 0") {
		t.Errorf("expected 'Left wins: 0', got:\n%s", output)
	}
}

func TestGenerateComparison_RightWinsAll(t *testing.T) {
	leftRuns := []*store.Run{
		makeRun("l1", "left", "tier1/task-a", false, 0, 8, 0, 5, 0.0, 60),
	}
	rightRuns := []*store.Run{
		makeRun("r1", "right", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 90),
	}

	var buf bytes.Buffer
	err := GenerateComparison(&buf, leftRuns, rightRuns, "left", "right")
	if err != nil {
		t.Fatalf("GenerateComparison: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Right wins: 1") {
		t.Errorf("expected 'Right wins: 1', got:\n%s", output)
	}
	if !strings.Contains(output, "Left wins: 0") {
		t.Errorf("expected 'Left wins: 0', got:\n%s", output)
	}
}

func TestGenerateComparison_DisjointTasks(t *testing.T) {
	// Left has task-a only, right has task-b only.
	leftRuns := []*store.Run{
		makeRun("l1", "left", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 60),
	}
	rightRuns := []*store.Run{
		makeRun("r1", "right", "tier1/task-b", true, 8, 8, 5, 5, 0.8, 90),
	}

	var buf bytes.Buffer
	err := GenerateComparison(&buf, leftRuns, rightRuns, "left", "right")
	if err != nil {
		t.Fatalf("GenerateComparison: %v", err)
	}

	output := buf.String()
	// Both tasks should appear in the comparison.
	if !strings.Contains(output, "tier1/task-a") {
		t.Error("expected task-a in output")
	}
	if !strings.Contains(output, "tier1/task-b") {
		t.Error("expected task-b in output")
	}
	// task-a: left=1.0, right=0.0 -> Left wins.
	// task-b: left=0.0, right=0.8 -> Right wins.
	if !strings.Contains(output, "Left wins: 1") {
		t.Errorf("expected 'Left wins: 1', got:\n%s", output)
	}
	if !strings.Contains(output, "Right wins: 1") {
		t.Errorf("expected 'Right wins: 1', got:\n%s", output)
	}
}

func TestGenerateComparison_WallTimeDelta(t *testing.T) {
	leftRuns := []*store.Run{
		makeRun("l1", "left", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 60),
	}
	rightRuns := []*store.Run{
		makeRun("r1", "right", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 600),
	}

	var buf bytes.Buffer
	err := GenerateComparison(&buf, leftRuns, rightRuns, "left", "right")
	if err != nil {
		t.Fatalf("GenerateComparison: %v", err)
	}

	output := buf.String()
	// Wall time ratio: 600/60 = 10x.
	if !strings.Contains(output, "Avg Wall Time") {
		t.Error("expected 'Avg Wall Time' in output")
	}
	if !strings.Contains(output, "+10.0x") {
		t.Errorf("expected '+10.0x' wall time delta, got:\n%s", output)
	}
}

func TestGenerateComparison_NoWallTime(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	// Create runs without wall time data.
	leftRuns := []*store.Run{
		{
			ID: "l1", Tag: "left", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: ptrBool(true), CorrectnessScore: ptrFloat(1.0),
		},
	}
	rightRuns := []*store.Run{
		{
			ID: "r1", Tag: "right", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: ptrBool(true), CorrectnessScore: ptrFloat(1.0),
		},
	}

	var buf bytes.Buffer
	err := GenerateComparison(&buf, leftRuns, rightRuns, "left", "right")
	if err != nil {
		t.Fatalf("GenerateComparison: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Avg Wall Time") {
		t.Errorf("expected no 'Avg Wall Time' when no data, got:\n%s", output)
	}
}

func TestGenerateComparison_MultipleRunsPerTask(t *testing.T) {
	// Left: task-a has 2 runs, average score = (1.0 + 0.8) / 2 = 0.9.
	leftRuns := []*store.Run{
		makeRun("l1", "left", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 60),
		makeRun("l2", "left", "tier1/task-a", true, 6, 8, 4, 5, 0.8, 70),
	}
	// Right: task-a has 1 run, score = 0.95.
	rightRuns := []*store.Run{
		makeRun("r1", "right", "tier1/task-a", true, 8, 8, 5, 5, 0.95, 90),
	}

	var buf bytes.Buffer
	err := GenerateComparison(&buf, leftRuns, rightRuns, "left", "right")
	if err != nil {
		t.Fatalf("GenerateComparison: %v", err)
	}

	output := buf.String()
	// Left avg correctness = 0.9 (averaged), right = 0.95 -> Right should win.
	if !strings.Contains(output, "Right wins: 1") {
		t.Errorf("expected 'Right wins: 1' (avg 0.9 vs 0.95), got:\n%s", output)
	}
}

func TestGenerateComparison_DeltaPassRate(t *testing.T) {
	leftRuns := []*store.Run{
		makeRun("l1", "left", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 60),
		makeRun("l2", "left", "tier1/task-b", true, 8, 8, 5, 5, 1.0, 60),
	}
	rightRuns := []*store.Run{
		makeRun("r1", "right", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 90),
		makeRun("r2", "right", "tier1/task-b", false, 0, 8, 0, 5, 0.0, 90),
	}

	var buf bytes.Buffer
	err := GenerateComparison(&buf, leftRuns, rightRuns, "left", "right")
	if err != nil {
		t.Fatalf("GenerateComparison: %v", err)
	}

	output := buf.String()
	// Left pass rate = 100%, right = 50%, delta = -50.0%.
	if !strings.Contains(output, "-50.0%") {
		t.Errorf("expected '-50.0%%' delta pass rate, got:\n%s", output)
	}
}

// TestGenerateComparison_PerTierComparison verifies that Per-Tier Comparison
// section appears when runs span multiple tiers.
func TestGenerateComparison_PerTierComparison(t *testing.T) {
	leftRuns := []*store.Run{
		makeRun("l1", "left", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 60),
		makeRun("l2", "left", "tier2/task-b", true, 6, 8, 3, 5, 0.6, 90),
	}
	// Override tier for the second run.
	leftRuns[1].Tier = 2

	rightRuns := []*store.Run{
		makeRun("r1", "right", "tier1/task-a", true, 8, 8, 5, 5, 0.9, 70),
		makeRun("r2", "right", "tier2/task-b", true, 8, 8, 5, 5, 1.0, 80),
	}
	rightRuns[1].Tier = 2

	var buf bytes.Buffer
	err := GenerateComparison(&buf, leftRuns, rightRuns, "left", "right")
	if err != nil {
		t.Fatalf("GenerateComparison: %v", err)
	}

	output := buf.String()
	// Should contain the Per-Tier Comparison section.
	if !strings.Contains(output, "Per-Tier Comparison") {
		t.Errorf("expected 'Per-Tier Comparison' section, got:\n%s", output)
	}
	// Should contain T1 and T2.
	if !strings.Contains(output, "T1") {
		t.Errorf("expected T1 in per-tier comparison, got:\n%s", output)
	}
	if !strings.Contains(output, "T2") {
		t.Errorf("expected T2 in per-tier comparison, got:\n%s", output)
	}
}

// TestFormatDuration verifies the formatDuration helper.
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		secs float64
		want string
	}{
		{0, "-"},
		{60, "1m0s"},
		{77, "1m17s"},
		{893, "14m53s"},
		{3600, "1h0m0s"},
	}
	for _, tc := range tests {
		got := formatDuration(tc.secs)
		if got != tc.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tc.secs, got, tc.want)
		}
	}
}

// TestFormatDeltaPercent verifies the formatDeltaPercent helper.
func TestFormatDeltaPercent(t *testing.T) {
	tests := []struct {
		delta float64
		want  string
	}{
		{20.0, "+20.0%"},
		{-5.0, "-5.0%"},
		{0, "+0.0%"},
	}
	for _, tc := range tests {
		got := formatDeltaPercent(tc.delta)
		if got != tc.want {
			t.Errorf("formatDeltaPercent(%v) = %q, want %q", tc.delta, got, tc.want)
		}
	}
}

// TestFormatMultiplier verifies the formatMultiplier helper.
func TestFormatMultiplier(t *testing.T) {
	tests := []struct {
		ratio float64
		want  string
	}{
		{10.0, "+10.0x"},
		{1.0, "+1.0x"},
		{0.5, "0.5x"},
	}
	for _, tc := range tests {
		got := formatMultiplier(tc.ratio)
		if got != tc.want {
			t.Errorf("formatMultiplier(%v) = %q, want %q", tc.ratio, got, tc.want)
		}
	}
}

// TestGenerateComparison_WithRubric verifies that the rubric section appears
// in comparison reports when runs have rubric scores.
func TestGenerateComparison_WithRubric(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)

	leftRuns := []*store.Run{
		{
			ID: "l-rubric-1", Tag: "left", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: ptrBool(true), CorrectnessScore: ptrFloat(1.0),
			RubricCorrectness:     ptrFloat(4.0),
			RubricReadability:     ptrFloat(3.0),
			RubricSimplicity:      ptrFloat(4.0),
			RubricRobustness:      ptrFloat(3.0),
			RubricMinimality:      ptrFloat(5.0),
			RubricMaintainability: ptrFloat(4.0),
			RubricGoIdioms:        ptrFloat(4.0),
			RubricComposite:       ptrFloat(3.85),
		},
	}
	rightRuns := []*store.Run{
		{
			ID: "r-rubric-1", Tag: "right", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: ptrBool(true), CorrectnessScore: ptrFloat(1.0),
			RubricCorrectness:     ptrFloat(5.0),
			RubricReadability:     ptrFloat(4.0),
			RubricSimplicity:      ptrFloat(5.0),
			RubricRobustness:      ptrFloat(4.0),
			RubricMinimality:      ptrFloat(5.0),
			RubricMaintainability: ptrFloat(5.0),
			RubricGoIdioms:        ptrFloat(5.0),
			RubricComposite:       ptrFloat(4.75),
		},
	}

	var buf bytes.Buffer
	err := GenerateComparison(&buf, leftRuns, rightRuns, "left", "right")
	if err != nil {
		t.Fatalf("GenerateComparison: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Code Quality (LLM Judge)") {
		t.Errorf("expected 'Code Quality (LLM Judge)' section, got:\n%s", output)
	}
	if !strings.Contains(output, "Composite") {
		t.Errorf("expected 'Composite' in rubric section, got:\n%s", output)
	}
}

// TestGenerateComparison_NoRubric verifies that the rubric section is omitted
// when neither side has rubric scores.
func TestGenerateComparison_NoRubric(t *testing.T) {
	leftRuns := []*store.Run{
		makeRun("l1", "left", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 60),
	}
	rightRuns := []*store.Run{
		makeRun("r1", "right", "tier1/task-a", true, 8, 8, 5, 5, 0.9, 90),
	}

	var buf bytes.Buffer
	err := GenerateComparison(&buf, leftRuns, rightRuns, "left", "right")
	if err != nil {
		t.Fatalf("GenerateComparison: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Code Quality (LLM Judge)") {
		t.Errorf("expected no rubric section when no rubric data, got:\n%s", output)
	}
}

// TestGenerateComparison_WithTokensAndCost verifies that token and cost rows
// appear in the comparison report when runs have token data.
func TestGenerateComparison_WithTokensAndCost(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)

	inputTok1 := 10000
	outputTok1 := 5000
	totalTok1 := 15000
	cost1 := 0.105 // 10000/1M*3 + 5000/1M*15 = 0.03 + 0.075
	inputTok2 := 100000
	outputTok2 := 50000
	totalTok2 := 150000
	cost2 := 1.05 // 100000/1M*3 + 50000/1M*15 = 0.30 + 0.75

	leftRuns := []*store.Run{
		{
			ID: "l-tok-1", Tag: "left", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: ptrBool(true), CorrectnessScore: ptrFloat(1.0),
			InputTokens: &inputTok1, OutputTokens: &outputTok1, TotalTokens: &totalTok1,
			CostUSD: &cost1,
		},
	}
	rightRuns := []*store.Run{
		{
			ID: "r-tok-1", Tag: "right", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: ptrBool(true), CorrectnessScore: ptrFloat(1.0),
			InputTokens: &inputTok2, OutputTokens: &outputTok2, TotalTokens: &totalTok2,
			CostUSD: &cost2,
		},
	}

	var buf bytes.Buffer
	err := GenerateComparison(&buf, leftRuns, rightRuns, "left", "right")
	if err != nil {
		t.Fatalf("GenerateComparison: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Avg Tokens") {
		t.Errorf("expected 'Avg Tokens' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Est. Cost/Task") {
		t.Errorf("expected 'Est. Cost/Task' in output, got:\n%s", output)
	}
	// Left: 15,000 tokens, Right: 150,000 tokens -> 10x ratio
	if !strings.Contains(output, "15,000") {
		t.Errorf("expected '15,000' for left avg tokens, got:\n%s", output)
	}
	if !strings.Contains(output, "150,000") {
		t.Errorf("expected '150,000' for right avg tokens, got:\n%s", output)
	}
	if !strings.Contains(output, "+10.0x") {
		t.Errorf("expected '+10.0x' token delta, got:\n%s", output)
	}
	// Left cost: 10000/1M*3 + 5000/1M*15 = 0.03 + 0.075 = $0.105 (rounds to $0.10)
	if !strings.Contains(output, "$0.10") {
		t.Errorf("expected '$0.10' for left cost, got:\n%s", output)
	}
	// Right cost: 100000/1M*3 + 50000/1M*15 = 0.30 + 0.75 = $1.05
	if !strings.Contains(output, "$1.05") {
		t.Errorf("expected '$1.05' for right cost, got:\n%s", output)
	}
}

// TestGenerateComparison_NoTokens verifies that token/cost rows are omitted
// when runs have no token data.
func TestGenerateComparison_NoTokens(t *testing.T) {
	leftRuns := []*store.Run{
		makeRun("l1", "left", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 60),
	}
	rightRuns := []*store.Run{
		makeRun("r1", "right", "tier1/task-a", true, 8, 8, 5, 5, 0.9, 90),
	}

	var buf bytes.Buffer
	err := GenerateComparison(&buf, leftRuns, rightRuns, "left", "right")
	if err != nil {
		t.Fatalf("GenerateComparison: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Avg Tokens") {
		t.Errorf("expected no 'Avg Tokens' when no token data, got:\n%s", output)
	}
	if strings.Contains(output, "Est. Cost/Task") {
		t.Errorf("expected no 'Est. Cost/Task' when no token data, got:\n%s", output)
	}
}

// TestEstimateCost_ViaMetrics verifies cost estimation using the unified metrics.EstimateCost.
func TestEstimateCost_ViaMetrics(t *testing.T) {
	// 1M input + 1M output at $3/$15 = $18
	cost := metrics.EstimateCost(1_000_000, 1_000_000, 3.0, 15.0)
	if cost != 18.0 {
		t.Errorf("EstimateCost(1M, 1M, 3, 15) = %v, want 18.0", cost)
	}

	// 0 tokens = $0
	cost = metrics.EstimateCost(0, 0, 3.0, 15.0)
	if cost != 0.0 {
		t.Errorf("EstimateCost(0, 0, 3, 15) = %v, want 0.0", cost)
	}

	// 10000 input + 5000 output = 0.03 + 0.075 = $0.105
	cost = metrics.EstimateCost(10000, 5000, 3.0, 15.0)
	expected := 0.105
	if cost < expected-0.001 || cost > expected+0.001 {
		t.Errorf("EstimateCost(10000, 5000, 3, 15) = %v, want ~%v", cost, expected)
	}
}

// TestFormatTokens verifies the formatTokens helper.
func TestFormatTokens(t *testing.T) {
	tests := []struct {
		tokens float64
		want   string
	}{
		{0, "-"},
		{999, "999"},
		{1000, "1,000"},
		{45231, "45,231"},
		{152800, "152,800"},
		{1234567, "1,234,567"},
	}
	for _, tc := range tests {
		got := formatTokens(tc.tokens)
		if got != tc.want {
			t.Errorf("formatTokens(%v) = %q, want %q", tc.tokens, got, tc.want)
		}
	}
}

// TestFormatCost verifies the formatCost helper.
func TestFormatCost(t *testing.T) {
	tests := []struct {
		cost float64
		want string
	}{
		{0, "-"},
		{0.28, "$0.28"},
		{0.95, "$0.95"},
		{18.0, "$18.00"},
	}
	for _, tc := range tests {
		got := formatCost(tc.cost)
		if got != tc.want {
			t.Errorf("formatCost(%v) = %q, want %q", tc.cost, got, tc.want)
		}
	}
}

// TestGenerateComparison_E2ECasesColumn verifies that the P15 E2E Cases column
// appears in the Per-Task Comparison table with the correct max L4 total.
func TestGenerateComparison_E2ECasesColumn(t *testing.T) {
	leftRuns := []*store.Run{
		makeRun("l1", "left", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 60),
	}
	rightRuns := []*store.Run{
		makeRun("r1", "right", "tier1/task-a", true, 8, 8, 3, 5, 0.8, 90),
	}

	var buf bytes.Buffer
	err := GenerateComparison(&buf, leftRuns, rightRuns, "left", "right")
	if err != nil {
		t.Fatalf("GenerateComparison: %v", err)
	}

	output := buf.String()
	// Should contain the E2E Cases column header.
	if !strings.Contains(output, "E2E Cases") {
		t.Errorf("expected 'E2E Cases' column header, got:\n%s", output)
	}
	// Task-a has L4Total=5 on both sides -> E2E Cases should show "5".
	// The row should contain "| 5 |" for E2E Cases.
	if !strings.Contains(output, "| 5 |") {
		t.Errorf("expected '| 5 |' in E2E Cases column, got:\n%s", output)
	}
}

// TestGenerateComparison_E2ECasesNoTests verifies that E2E Cases shows "-" when
// there are no E2E tests.
func TestGenerateComparison_E2ECasesNoTests(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	// Create runs without L4 data.
	leftRuns := []*store.Run{
		{
			ID: "l1", Tag: "left", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: ptrBool(true), CorrectnessScore: ptrFloat(1.0),
		},
	}
	rightRuns := []*store.Run{
		{
			ID: "r1", Tag: "right", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: ptrBool(true), CorrectnessScore: ptrFloat(1.0),
		},
	}

	var buf bytes.Buffer
	err := GenerateComparison(&buf, leftRuns, rightRuns, "left", "right")
	if err != nil {
		t.Fatalf("GenerateComparison: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "E2E Cases") {
		t.Errorf("expected 'E2E Cases' column header, got:\n%s", output)
	}
	// Should contain "| - |" for E2E Cases when no L4 data.
	if !strings.Contains(output, "| - |") {
		t.Errorf("expected '| - |' for no E2E data, got:\n%s", output)
	}
}

// TestGenerateComparisonWithPairwise verifies that pairwise results appear in the report.
func TestGenerateComparisonWithPairwise(t *testing.T) {
	leftRuns := []*store.Run{
		makeRun("l1", "left", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 60),
	}
	rightRuns := []*store.Run{
		makeRun("r1", "right", "tier1/task-a", true, 8, 8, 5, 5, 0.9, 90),
	}

	pairwiseResults := []TaskPairwise{
		{
			TaskID:             "tier1/task-a",
			Winner:             "Left",
			PositionConsistent: true,
			Dimensions: map[string]string{
				"correctness": "left",
				"readability": "tie",
			},
		},
	}

	var buf bytes.Buffer
	err := GenerateComparisonWithPairwise(&buf, leftRuns, rightRuns, "left", "right", pairwiseResults)
	if err != nil {
		t.Fatalf("GenerateComparisonWithPairwise: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Pairwise Comparison") {
		t.Errorf("expected 'Pairwise Comparison' section, got:\n%s", output)
	}
	if !strings.Contains(output, "tier1/task-a") {
		t.Errorf("expected task-a in pairwise section, got:\n%s", output)
	}
	if !strings.Contains(output, "Left") {
		t.Errorf("expected 'Left' winner in pairwise section, got:\n%s", output)
	}
	if !strings.Contains(output, "Yes") {
		t.Errorf("expected 'Yes' for position consistent, got:\n%s", output)
	}
	// Summary line.
	if !strings.Contains(output, "Left wins: 1") {
		t.Errorf("expected pairwise 'Left wins: 1', got:\n%s", output)
	}
}

// TestGenerateComparisonWithPairwise_NoPairwise verifies pairwise section is omitted
// when no pairwise results are provided.
func TestGenerateComparisonWithPairwise_NoPairwise(t *testing.T) {
	leftRuns := []*store.Run{
		makeRun("l1", "left", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 60),
	}
	rightRuns := []*store.Run{
		makeRun("r1", "right", "tier1/task-a", true, 8, 8, 5, 5, 0.9, 90),
	}

	var buf bytes.Buffer
	err := GenerateComparisonWithPairwise(&buf, leftRuns, rightRuns, "left", "right", nil)
	if err != nil {
		t.Fatalf("GenerateComparisonWithPairwise: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Pairwise Comparison") {
		t.Errorf("expected no pairwise section when no results, got:\n%s", output)
	}
}
