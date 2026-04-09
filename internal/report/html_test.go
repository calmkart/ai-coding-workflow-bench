package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/calmkart/ai-coding-workflow-bench/internal/store"
)

func TestGenerateSummaryHTML_Basic(t *testing.T) {
	now := time.Now()
	finished := now.Add(2 * time.Minute)
	l1 := true
	l2p, l2t := 8, 8
	l3 := 0
	l4p, l4t := 5, 5
	cs := 0.95

	runs := []*store.Run{
		{
			ID: "test-html-001", Tag: "smoke", Workflow: "vanilla",
			TaskID: "tier1/fix-handler-bug", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, L2UtPassed: &l2p, L2UtTotal: &l2t, L3LintIssues: &l3,
			L4E2EPassed: &l4p, L4E2ETotal: &l4t, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummaryHTML(&buf, runs); err != nil {
		t.Fatalf("GenerateSummaryHTML: %v", err)
	}

	output := buf.String()

	// Verify it is valid HTML.
	if !strings.Contains(output, "<html") {
		t.Error("expected <html> tag in output")
	}
	if !strings.Contains(output, "</html>") {
		t.Error("expected </html> closing tag")
	}

	// Verify report title.
	if !strings.Contains(output, "Benchmark Report: smoke") {
		t.Error("expected report header")
	}

	// Verify table structure.
	if !strings.Contains(output, "<table>") {
		t.Error("expected <table> tag")
	}
	if !strings.Contains(output, "<th>") {
		t.Error("expected <th> tag")
	}
	if !strings.Contains(output, "<td>") {
		t.Error("expected <td> tag")
	}

	// Verify task data.
	if !strings.Contains(output, "tier1/fix-handler-bug") {
		t.Error("expected task ID in report")
	}
	if !strings.Contains(output, "100.0%") {
		t.Error("expected 100% pass rate")
	}

	// Verify L1 PASS is styled.
	if !strings.Contains(output, `class="pass"`) {
		t.Error("expected pass CSS class")
	}
	if !strings.Contains(output, "PASS") {
		t.Error("expected PASS text")
	}

	// Verify 8/8 in L2 column.
	if !strings.Contains(output, "8/8") {
		t.Error("expected 8/8 in L2 column")
	}

	// Verify 5/5 in L4 column.
	if !strings.Contains(output, "5/5") {
		t.Error("expected 5/5 in L4 column")
	}

	// Verify inline CSS is present.
	if !strings.Contains(output, "<style>") {
		t.Error("expected inline <style> tag")
	}
	if !strings.Contains(output, "border-collapse") {
		t.Error("expected CSS border-collapse rule")
	}
}

func TestGenerateSummaryHTML_Empty(t *testing.T) {
	var buf bytes.Buffer
	err := GenerateSummaryHTML(&buf, nil)
	if err == nil {
		t.Error("expected error for empty runs")
	}
}

func TestGenerateSummaryHTML_L1FailHighlighted(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1Fail := false
	cs := 0.0

	runs := []*store.Run{
		{
			ID: "html-fail-1", Tag: "fail-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1Fail, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummaryHTML(&buf, runs); err != nil {
		t.Fatalf("GenerateSummaryHTML: %v", err)
	}

	output := buf.String()

	// L1=FAIL should have fail class for red styling.
	if !strings.Contains(output, `class="fail"`) {
		t.Error("expected fail CSS class for L1=FAIL")
	}
	if !strings.Contains(output, "FAIL") {
		t.Error("expected FAIL text")
	}
}

func TestGenerateSummaryHTML_PerTierColorCoding(t *testing.T) {
	now := time.Now()
	finished := now.Add(2 * time.Minute)
	l1 := true
	l2p, l2t := 8, 8
	l3 := 0
	l4p, l4t := 5, 5
	cs := 1.0
	l4pFail, l4tFail := 0, 5
	csFail := 0.0

	runs := []*store.Run{
		// Tier 1: all pass -> green (>80%).
		{
			ID: "tier-html-1", Tag: "tier-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, L2UtPassed: &l2p, L2UtTotal: &l2t, L3LintIssues: &l3,
			L4E2EPassed: &l4p, L4E2ETotal: &l4t, CorrectnessScore: &cs,
		},
		// Tier 2: fail -> red (<50%).
		{
			ID: "tier-html-2", Tag: "tier-test", Workflow: "vanilla",
			TaskID: "tier2/task-b", Tier: 2, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, L2UtPassed: &l2p, L2UtTotal: &l2t, L3LintIssues: &l3,
			L4E2EPassed: &l4pFail, L4E2ETotal: &l4tFail, CorrectnessScore: &csFail,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummaryHTML(&buf, runs); err != nil {
		t.Fatalf("GenerateSummaryHTML: %v", err)
	}

	output := buf.String()

	// Verify tier section exists.
	if !strings.Contains(output, "Per-Tier Summary") {
		t.Error("expected Per-Tier Summary section")
	}

	// T1 should have green class (100% pass rate).
	if !strings.Contains(output, "rate-green") {
		t.Error("expected rate-green CSS class for T1 (100% pass rate)")
	}

	// T2 should have red class (0% pass rate).
	if !strings.Contains(output, "rate-red") {
		t.Error("expected rate-red CSS class for T2 (0% pass rate)")
	}
}

func TestGenerateSummaryHTML_WithRubricBarChart(t *testing.T) {
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
			ID: "html-rubric-1", Tag: "rubric-test", Workflow: "vanilla",
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
	if err := GenerateSummaryHTML(&buf, runs); err != nil {
		t.Fatalf("GenerateSummaryHTML: %v", err)
	}

	output := buf.String()

	// Verify rubric section exists with bar chart elements.
	if !strings.Contains(output, "Code Quality (LLM Judge)") {
		t.Error("expected Code Quality section")
	}
	if !strings.Contains(output, "bar-container") {
		t.Error("expected bar-container CSS class for bar chart")
	}
	if !strings.Contains(output, `class="bar"`) {
		t.Error("expected bar CSS class for bar chart")
	}
	// Correctness is 4.0/5 = 80%, check the bar width.
	if !strings.Contains(output, "80%") {
		t.Error("expected 80% bar width for Correctness=4.0")
	}
	// Minimality is 5.0/5 = 100%.
	if !strings.Contains(output, "100%") {
		t.Error("expected 100% bar width for Minimality=5.0")
	}
}

func TestGenerateSummaryHTML_NoRubric(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	l1 := true
	cs := 0.9

	runs := []*store.Run{
		{
			ID: "html-norubric", Tag: "norubric-test", Workflow: "vanilla",
			TaskID: "tier1/task-a", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed", StartedAt: now, FinishedAt: &finished,
			L1Build: &l1, CorrectnessScore: &cs,
		},
	}

	var buf bytes.Buffer
	if err := GenerateSummaryHTML(&buf, runs); err != nil {
		t.Fatalf("GenerateSummaryHTML: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Code Quality (LLM Judge)") {
		t.Error("expected no rubric section when no rubric data")
	}
}

func TestGenerateComparisonHTML_Basic(t *testing.T) {
	leftRuns := []*store.Run{
		makeRun("l-html-1", "vanilla-tag", "tier1/fix-handler-bug", true, 8, 8, 5, 5, 1.0, 64),
		makeRun("l-html-2", "vanilla-tag", "tier1/fix-status-code", true, 8, 8, 5, 5, 1.0, 63),
	}
	rightRuns := []*store.Run{
		makeRun("r-html-1", "v4-tag", "tier1/fix-handler-bug", false, 2, 8, 0, 5, 0.30, 912),
		makeRun("r-html-2", "v4-tag", "tier1/fix-status-code", true, 8, 8, 5, 5, 1.0, 750),
	}

	var buf bytes.Buffer
	err := GenerateComparisonHTML(&buf, leftRuns, rightRuns, "vanilla-tag", "v4-tag")
	if err != nil {
		t.Fatalf("GenerateComparisonHTML: %v", err)
	}

	output := buf.String()

	// Verify HTML structure.
	if !strings.Contains(output, "<html") {
		t.Error("expected <html> tag")
	}
	if !strings.Contains(output, "</html>") {
		t.Error("expected </html> closing tag")
	}

	// Verify header.
	if !strings.Contains(output, "Comparison: vanilla-tag vs v4-tag") {
		t.Error("expected comparison header")
	}

	// Verify table structure.
	if !strings.Contains(output, "<table>") {
		t.Error("expected <table> tag")
	}

	// Verify tags appear.
	if !strings.Contains(output, "vanilla-tag") {
		t.Error("expected left tag in output")
	}
	if !strings.Contains(output, "v4-tag") {
		t.Error("expected right tag in output")
	}

	// Verify pass/fail icons are rendered as HTML entities.
	if !strings.Contains(output, "&#x2705;") || !strings.Contains(output, "&#x274C;") {
		t.Error("expected pass/fail HTML entities (checkmark and X)")
	}

	// Verify delta classes are present.
	if !strings.Contains(output, "delta-") {
		t.Error("expected delta CSS classes in output")
	}

	// Verify summary section.
	if !strings.Contains(output, "Left wins:") {
		t.Error("expected summary section with Left wins")
	}

	// Verify inline CSS.
	if !strings.Contains(output, "<style>") {
		t.Error("expected inline <style> tag")
	}
}

func TestGenerateComparisonHTML_EmptyLeft(t *testing.T) {
	rightRuns := []*store.Run{
		makeRun("r1", "right", "tier1/fix", true, 8, 8, 5, 5, 1.0, 60),
	}

	var buf bytes.Buffer
	err := GenerateComparisonHTML(&buf, nil, rightRuns, "empty", "right")
	if err == nil {
		t.Fatal("expected error for empty left runs")
	}
}

func TestGenerateComparisonHTML_EmptyRight(t *testing.T) {
	leftRuns := []*store.Run{
		makeRun("l1", "left", "tier1/fix", true, 8, 8, 5, 5, 1.0, 60),
	}

	var buf bytes.Buffer
	err := GenerateComparisonHTML(&buf, leftRuns, nil, "left", "empty")
	if err == nil {
		t.Fatal("expected error for empty right runs")
	}
}

func TestPassRateClass(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"100.0%", "rate-green"},
		{"85.0%", "rate-green"},
		{"80.1%", "rate-green"},
		{"80.0%", "rate-yellow"},
		{"65.0%", "rate-yellow"},
		{"50.0%", "rate-yellow"},
		{"49.9%", "rate-red"},
		{"0.0%", "rate-red"},
		{"-", ""},
		{"", ""},
	}
	for _, tc := range tests {
		got := passRateClass(tc.input)
		if got != tc.want {
			t.Errorf("passRateClass(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestDeltaClass(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"+20.0%", "delta-pos"},
		{"-5.0%", "delta-neg"},
		{"+0.0%", "delta-zero"},
		{"+0.00", "delta-zero"},
		{"-", "delta-zero"},
		{"", "delta-zero"},
	}
	for _, tc := range tests {
		got := deltaClass(tc.input)
		if got != tc.want {
			t.Errorf("deltaClass(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestDeltaClassFloat(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0.5, "delta-pos"},
		{-0.3, "delta-neg"},
		{0.0, "delta-zero"},
	}
	for _, tc := range tests {
		got := deltaClassFloat(tc.input)
		if got != tc.want {
			t.Errorf("deltaClassFloat(%v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
