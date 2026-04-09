// Package report generates benchmark reports from stored results.
package report

import (
	_ "embed"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/calmkart/ai-coding-workflow-bench/internal/metrics"
	"github.com/calmkart/ai-coding-workflow-bench/internal/store"
)

//go:embed templates/compare.md.tmpl
var compareTemplate string

// TaskComparison holds per-task comparison data for the compare report template.
type TaskComparison struct {
	TaskID     string
	LeftPass   string  // checkmark or X
	RightPass  string  // checkmark or X
	LeftScore  float64 // average correctness score
	RightScore float64 // average correctness score
	LeftTime   string  // formatted average wall time
	RightTime  string  // formatted average wall time
	Winner     string  // "Left", "Right", or "Tie"
	E2ECases   string  // P15: max E2E total across left/right runs for this task (e.g. "5")
}

// TierComparison holds per-tier comparison data for the compare report template.
type TierComparison struct {
	Tier              int
	LeftPassRate      string
	RightPassRate     string
	LeftAvgCorr       string
	RightAvgCorr      string
	DeltaCorrectness  string
	Significant       bool // true if CI of left and right pass rates don't overlap
}

// CompareData holds all data for the comparison report template.
type CompareData struct {
	LeftTag  string
	RightTag string

	// Dashes for table header alignment (match tag name length).
	LeftTagDashes  string
	RightTagDashes string

	// Overall metrics.
	LeftPassRate       float64
	RightPassRate      float64
	LeftAvgCorrectness float64
	RightAvgCorrectness float64
	LeftAvgWallTime    string
	RightAvgWallTime   string
	LeftAvgL2          float64
	RightAvgL2         float64

	// CI-enhanced pass rates (P1).
	LeftPassRateCI  string // "95.0% [87.2-98.6]"
	RightPassRateCI string

	// Delta strings.
	DeltaPassRate    string // includes * if significant
	DeltaCorrectness string
	DeltaWallTime    string
	DeltaL2          string

	// Statistical significance (P1).
	PassRateSignificant bool   // true if left/right CIs don't overlap
	LowSampleWarning   string // non-empty if K < 5 for either side

	// Token and cost metrics.
	HasTokens     bool
	LeftAvgTokens string
	RightAvgTokens string
	DeltaTokens   string
	LeftAvgCost   string
	RightAvgCost  string
	DeltaCost     string

	// Feature flags for conditional template sections.
	HasWallTime bool
	HasL2Tests  bool

	// Per-tier comparisons.
	TierComparisons []TierComparison

	// Per-task comparisons.
	Tasks []TaskComparison

	// Summary counts.
	LeftWins  int
	RightWins int
	Ties      int

	// Rubric comparison.
	HasRubric      bool
	LeftRubric     RubricSummary
	RightRubric    RubricSummary
	DeltaComposite string

	// P17: Pairwise comparison results (populated externally when --pairwise flag is used).
	HasPairwise      bool
	PairwiseResults  []TaskPairwise
	PairwiseSummary  PairwiseSummaryData
}

// TaskPairwise holds per-task pairwise comparison data.
type TaskPairwise struct {
	TaskID             string
	Winner             string // "Left", "Right", "Tie"
	PositionConsistent bool
	Dimensions         map[string]string // dimension -> "Left"|"Right"|"Tie"
}

// PairwiseSummaryData holds aggregate pairwise comparison counts.
type PairwiseSummaryData struct {
	LeftWins  int
	RightWins int
	Ties      int
}

// GenerateComparison renders a comparison report between two sets of runs (left vs right).
//
// It matches runs by task_id, computes per-task and overall deltas, and determines
// a winner per task based on correctness score. When correctness is equal, it is a tie.
func GenerateComparison(w io.Writer, leftRuns, rightRuns []*store.Run, leftTag, rightTag string) error {
	return GenerateComparisonWithPairwise(w, leftRuns, rightRuns, leftTag, rightTag, nil)
}

// GenerateComparisonWithPairwise renders a comparison report with optional pairwise results.
// If pairwiseResults is nil or empty, the pairwise section is omitted.
//
// @implements P17 (Pairwise comparison report rendering)
func GenerateComparisonWithPairwise(w io.Writer, leftRuns, rightRuns []*store.Run, leftTag, rightTag string, pairwiseResults []TaskPairwise) error {
	if len(leftRuns) == 0 {
		return fmt.Errorf("no runs found for left tag %q", leftTag)
	}
	if len(rightRuns) == 0 {
		return fmt.Errorf("no runs found for right tag %q", rightTag)
	}

	funcMap := template.FuncMap{
		"sub": func(a, b float64) float64 { return a - b },
	}
	tmpl, err := template.New("compare").Funcs(funcMap).Parse(compareTemplate)
	if err != nil {
		return fmt.Errorf("parse compare template: %w", err)
	}

	data := buildCompareData(leftRuns, rightRuns, leftTag, rightTag)

	// P17: Attach pairwise results if available.
	if len(pairwiseResults) > 0 {
		data.HasPairwise = true
		data.PairwiseResults = pairwiseResults
		var leftWins, rightWins, ties int
		for _, pw := range pairwiseResults {
			switch pw.Winner {
			case "Left":
				leftWins++
			case "Right":
				rightWins++
			default:
				ties++
			}
		}
		data.PairwiseSummary = PairwiseSummaryData{
			LeftWins:  leftWins,
			RightWins: rightWins,
			Ties:      ties,
		}
	}

	return tmpl.Execute(w, data)
}

// taskStats aggregates metrics for a single task across multiple runs.
type taskStats struct {
	passCount  int
	totalCount int
	totalScore float64
	totalWall  float64
	wallCount  int
	totalL2    int
	l2Count    int
	maxE2ETotal int // P15: max L4 E2E total across runs for this task
}

func (ts *taskStats) avgScore() float64 {
	if ts.totalCount == 0 {
		return 0
	}
	return ts.totalScore / float64(ts.totalCount)
}

func (ts *taskStats) avgWall() float64 {
	if ts.wallCount == 0 {
		return 0
	}
	return ts.totalWall / float64(ts.wallCount)
}

func (ts *taskStats) avgL2() float64 {
	if ts.l2Count == 0 {
		return 0
	}
	return float64(ts.totalL2) / float64(ts.l2Count)
}

func (ts *taskStats) passed() bool {
	return ts.totalCount > 0 && ts.passCount == ts.totalCount
}

// aggregateByTask groups runs by task_id and computes per-task aggregate stats.
func aggregateByTask(runs []*store.Run) map[string]*taskStats {
	m := make(map[string]*taskStats)
	for _, r := range runs {
		ts, ok := m[r.TaskID]
		if !ok {
			ts = &taskStats{}
			m[r.TaskID] = ts
		}
		ts.totalCount++

		// A run "passes" if L1 build succeeds and all L4 E2E tests pass.
		l1Pass := r.L1Build != nil && *r.L1Build
		l4AllPass := r.L4E2EPassed != nil && r.L4E2ETotal != nil &&
			*r.L4E2ETotal > 0 && *r.L4E2EPassed == *r.L4E2ETotal
		if l1Pass && l4AllPass {
			ts.passCount++
		}

		if r.CorrectnessScore != nil {
			ts.totalScore += *r.CorrectnessScore
		}
		if r.WallTimeSecs != nil {
			ts.totalWall += *r.WallTimeSecs
			ts.wallCount++
		}
		if r.L2UtTotal != nil {
			ts.totalL2 += *r.L2UtTotal
			ts.l2Count++
		}
		// P15: Track max E2E total for this task.
		if r.L4E2ETotal != nil && *r.L4E2ETotal > ts.maxE2ETotal {
			ts.maxE2ETotal = *r.L4E2ETotal
		}
	}
	return m
}


// overallStats computes aggregate stats across all runs.
type overallStats struct {
	passRate       float64
	avgCorrectness float64
	avgWallTime    float64
	hasWallTime    bool
	avgL2          float64
	hasL2          bool
	avgTokens      float64
	avgCost        float64
	hasTokens      bool
}

func computeOverall(runs []*store.Run) overallStats {
	var os overallStats
	if len(runs) == 0 {
		return os
	}

	var passCount int
	var totalCorr float64
	var totalWall float64
	var wallCount int
	var totalL2 int
	var l2Count int
	var totalTokens int64
	var totalCost float64
	var tokenCount int

	for _, r := range runs {
		l1Pass := r.L1Build != nil && *r.L1Build
		l4AllPass := r.L4E2EPassed != nil && r.L4E2ETotal != nil &&
			*r.L4E2ETotal > 0 && *r.L4E2EPassed == *r.L4E2ETotal
		if l1Pass && l4AllPass {
			passCount++
		}
		if r.CorrectnessScore != nil {
			totalCorr += *r.CorrectnessScore
		}
		if r.WallTimeSecs != nil {
			totalWall += *r.WallTimeSecs
			wallCount++
		}
		if r.L2UtTotal != nil {
			totalL2 += *r.L2UtTotal
			l2Count++
		}
		if r.TotalTokens != nil {
			totalTokens += int64(*r.TotalTokens)
			tokenCount++
			// P8: Use pre-computed CostUSD from the run (already calculated by runner
			// with correct per-model pricing from config) instead of hardcoded pricing.
			if r.CostUSD != nil {
				totalCost += *r.CostUSD
			}
		}
	}

	os.passRate = float64(passCount) / float64(len(runs)) * 100
	os.avgCorrectness = totalCorr / float64(len(runs))
	if wallCount > 0 {
		os.avgWallTime = totalWall / float64(wallCount)
		os.hasWallTime = true
	}
	if l2Count > 0 {
		os.avgL2 = float64(totalL2) / float64(l2Count)
		os.hasL2 = true
	}
	if tokenCount > 0 {
		os.avgTokens = float64(totalTokens) / float64(tokenCount)
		os.avgCost = totalCost / float64(tokenCount)
		os.hasTokens = true
	}
	return os
}

func buildCompareData(leftRuns, rightRuns []*store.Run, leftTag, rightTag string) CompareData {
	leftOverall := computeOverall(leftRuns)
	rightOverall := computeOverall(rightRuns)

	leftByTask := aggregateByTask(leftRuns)
	rightByTask := aggregateByTask(rightRuns)

	// Collect all unique task IDs, sorted.
	taskSet := make(map[string]bool)
	for id := range leftByTask {
		taskSet[id] = true
	}
	for id := range rightByTask {
		taskSet[id] = true
	}
	taskIDs := make([]string, 0, len(taskSet))
	for id := range taskSet {
		taskIDs = append(taskIDs, id)
	}
	sort.Strings(taskIDs)

	// Build per-task comparisons.
	var tasks []TaskComparison
	var leftWins, rightWins, ties int

	for _, id := range taskIDs {
		tc := TaskComparison{TaskID: id}

		leftTS := leftByTask[id]
		rightTS := rightByTask[id]

		// Left side.
		if leftTS != nil {
			if leftTS.passed() {
				tc.LeftPass = "\u2705"
			} else {
				tc.LeftPass = "\u274c"
			}
			tc.LeftScore = leftTS.avgScore()
			if leftTS.wallCount > 0 {
				tc.LeftTime = formatDuration(leftTS.avgWall())
			} else {
				tc.LeftTime = "-"
			}
		} else {
			tc.LeftPass = "-"
			tc.LeftTime = "-"
		}

		// Right side.
		if rightTS != nil {
			if rightTS.passed() {
				tc.RightPass = "\u2705"
			} else {
				tc.RightPass = "\u274c"
			}
			tc.RightScore = rightTS.avgScore()
			if rightTS.wallCount > 0 {
				tc.RightTime = formatDuration(rightTS.avgWall())
			} else {
				tc.RightTime = "-"
			}
		} else {
			tc.RightPass = "-"
			tc.RightTime = "-"
		}

		// P15: E2E Cases - show max L4 total across left and right runs.
		maxE2E := 0
		if leftTS != nil && leftTS.maxE2ETotal > maxE2E {
			maxE2E = leftTS.maxE2ETotal
		}
		if rightTS != nil && rightTS.maxE2ETotal > maxE2E {
			maxE2E = rightTS.maxE2ETotal
		}
		if maxE2E > 0 {
			tc.E2ECases = fmt.Sprintf("%d", maxE2E)
		} else {
			tc.E2ECases = "-"
		}

		// Winner determination: higher correctness wins; tie if equal.
		leftScore := 0.0
		rightScore := 0.0
		if leftTS != nil {
			leftScore = leftTS.avgScore()
		}
		if rightTS != nil {
			rightScore = rightTS.avgScore()
		}

		// Use a small epsilon for float comparison.
		const epsilon = 1e-9
		if leftScore-rightScore > epsilon {
			tc.Winner = "Left"
			leftWins++
		} else if rightScore-leftScore > epsilon {
			tc.Winner = "Right"
			rightWins++
		} else {
			tc.Winner = "Tie"
			ties++
		}

		tasks = append(tasks, tc)
	}

	// Build per-tier comparisons.
	tierComparisons := buildTierComparisons(leftRuns, rightRuns)

	// Build dashes for table header alignment.
	leftDashes := strings.Repeat("-", max(len(leftTag)+2, 7))
	rightDashes := strings.Repeat("-", max(len(rightTag)+2, 7))

	data := CompareData{
		LeftTag:  leftTag,
		RightTag: rightTag,

		LeftTagDashes:  leftDashes,
		RightTagDashes: rightDashes,

		LeftPassRate:        leftOverall.passRate,
		RightPassRate:       rightOverall.passRate,
		LeftAvgCorrectness:  leftOverall.avgCorrectness,
		RightAvgCorrectness: rightOverall.avgCorrectness,

		HasWallTime: leftOverall.hasWallTime || rightOverall.hasWallTime,
		HasL2Tests:  leftOverall.hasL2 || rightOverall.hasL2,
		HasTokens:   leftOverall.hasTokens || rightOverall.hasTokens,

		TierComparisons: tierComparisons,
		Tasks:           tasks,
		LeftWins:  leftWins,
		RightWins: rightWins,
		Ties:      ties,
	}

	// P1: Confidence intervals and significance for pass rates.
	leftPassCount := int(math.Round(leftOverall.passRate / 100 * float64(len(leftRuns))))
	rightPassCount := int(math.Round(rightOverall.passRate / 100 * float64(len(rightRuns))))
	data.LeftPassRateCI = metrics.FormatCI(leftPassCount, len(leftRuns))
	data.RightPassRateCI = metrics.FormatCI(rightPassCount, len(rightRuns))

	ll, lu := metrics.WilsonCI(leftPassCount, len(leftRuns))
	rl, ru := metrics.WilsonCI(rightPassCount, len(rightRuns))
	data.PassRateSignificant = !metrics.CIOverlaps(ll, lu, rl, ru)

	// P1: Low sample size warning.
	if metrics.IsLowSampleSize(len(leftRuns)) || metrics.IsLowSampleSize(len(rightRuns)) {
		data.LowSampleWarning = "Low sample size (K < 5): results may not be statistically reliable"
	}

	// Delta calculations.
	deltaPassRate := formatDeltaPercent(rightOverall.passRate - leftOverall.passRate)
	if data.PassRateSignificant {
		deltaPassRate += " *"
	}
	data.DeltaPassRate = deltaPassRate
	data.DeltaCorrectness = formatDeltaFloat(rightOverall.avgCorrectness - leftOverall.avgCorrectness)

	if data.HasWallTime {
		data.LeftAvgWallTime = formatDuration(leftOverall.avgWallTime)
		data.RightAvgWallTime = formatDuration(rightOverall.avgWallTime)
		if leftOverall.hasWallTime && rightOverall.hasWallTime && leftOverall.avgWallTime > 0 {
			ratio := rightOverall.avgWallTime / leftOverall.avgWallTime
			data.DeltaWallTime = formatMultiplier(ratio)
		} else {
			data.DeltaWallTime = "-"
		}
	}

	if data.HasL2Tests {
		data.LeftAvgL2 = leftOverall.avgL2
		data.RightAvgL2 = rightOverall.avgL2
		if leftOverall.hasL2 && rightOverall.hasL2 && leftOverall.avgL2 > 0 {
			ratio := rightOverall.avgL2 / leftOverall.avgL2
			data.DeltaL2 = formatMultiplier(ratio)
		} else {
			data.DeltaL2 = "-"
		}
	}

	if data.HasTokens {
		data.LeftAvgTokens = formatTokens(leftOverall.avgTokens)
		data.RightAvgTokens = formatTokens(rightOverall.avgTokens)
		data.LeftAvgCost = formatCost(leftOverall.avgCost)
		data.RightAvgCost = formatCost(rightOverall.avgCost)
		if leftOverall.hasTokens && rightOverall.hasTokens && leftOverall.avgTokens > 0 {
			ratio := rightOverall.avgTokens / leftOverall.avgTokens
			data.DeltaTokens = formatMultiplier(ratio)
		} else {
			data.DeltaTokens = "-"
		}
		if leftOverall.hasTokens && rightOverall.hasTokens && leftOverall.avgCost > 0 {
			ratio := rightOverall.avgCost / leftOverall.avgCost
			data.DeltaCost = formatMultiplier(ratio)
		} else {
			data.DeltaCost = "-"
		}
	}

	// Rubric comparison.
	leftRubric, leftHasRubric := computeRubricAvg(leftRuns)
	rightRubric, rightHasRubric := computeRubricAvg(rightRuns)
	if leftHasRubric || rightHasRubric {
		data.HasRubric = true
		data.LeftRubric = leftRubric
		data.RightRubric = rightRubric
		if leftHasRubric && rightHasRubric {
			data.DeltaComposite = formatDeltaFloat(rightRubric.Composite - leftRubric.Composite)
		} else {
			data.DeltaComposite = "-"
		}
	}

	return data
}

// computeRubricAvg calculates average rubric scores across runs that have rubric data.
func computeRubricAvg(runs []*store.Run) (RubricSummary, bool) {
	var sum RubricSummary
	var count int
	for _, r := range runs {
		if r.RubricComposite == nil {
			continue
		}
		count++
		if r.RubricCorrectness != nil {
			sum.Correctness += *r.RubricCorrectness
		}
		if r.RubricReadability != nil {
			sum.Readability += *r.RubricReadability
		}
		if r.RubricSimplicity != nil {
			sum.Simplicity += *r.RubricSimplicity
		}
		if r.RubricRobustness != nil {
			sum.Robustness += *r.RubricRobustness
		}
		if r.RubricMinimality != nil {
			sum.Minimality += *r.RubricMinimality
		}
		if r.RubricMaintainability != nil {
			sum.Maintainability += *r.RubricMaintainability
		}
		if r.RubricGoIdioms != nil {
			sum.GoIdioms += *r.RubricGoIdioms
		}
		sum.Composite += *r.RubricComposite
	}
	if count == 0 {
		return RubricSummary{}, false
	}
	n := float64(count)
	return RubricSummary{
		Correctness:     sum.Correctness / n,
		Readability:     sum.Readability / n,
		Simplicity:      sum.Simplicity / n,
		Robustness:      sum.Robustness / n,
		Minimality:      sum.Minimality / n,
		Maintainability: sum.Maintainability / n,
		GoIdioms:        sum.GoIdioms / n,
		Composite:       sum.Composite / n,
	}, true
}

// formatDuration formats seconds into a human-readable duration string (e.g. "1m17s").
func formatDuration(secs float64) string {
	if secs == 0 {
		return "-"
	}
	d := time.Duration(secs * float64(time.Second))
	d = d.Truncate(time.Second)
	return d.String()
}

// formatDeltaPercent formats a percentage delta with sign (e.g. "+20.0%", "-5.0%").
func formatDeltaPercent(delta float64) string {
	if delta >= 0 {
		return fmt.Sprintf("+%.1f%%", delta)
	}
	return fmt.Sprintf("%.1f%%", delta)
}

// formatDeltaFloat formats a float delta with sign (e.g. "+0.06", "-0.12").
func formatDeltaFloat(delta float64) string {
	if delta >= 0 {
		return fmt.Sprintf("+%.2f", delta)
	}
	return fmt.Sprintf("%.2f", delta)
}

// formatTokens formats an average token count as a comma-separated string (e.g. "45,231").
func formatTokens(tokens float64) string {
	n := int64(tokens + 0.5)
	if n == 0 {
		return "-"
	}
	// Simple comma formatting for integers.
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return strings.Join(parts, ",")
}

// formatCost formats a USD cost as "$0.28".
func formatCost(cost float64) string {
	if cost == 0 {
		return "-"
	}
	return fmt.Sprintf("$%.2f", cost)
}

// formatMultiplier formats a ratio as a multiplier string (e.g. "+11.5x", "1.0x").
func formatMultiplier(ratio float64) string {
	if math.IsNaN(ratio) || math.IsInf(ratio, 0) {
		return "-"
	}
	if ratio >= 1.0 {
		return fmt.Sprintf("+%.1fx", ratio)
	}
	return fmt.Sprintf("%.1fx", ratio)
}

// tierRunStats holds per-tier aggregated stats for a set of runs.
type tierRunStats struct {
	passCount  int
	totalCount int
	totalCorr  float64
}

// buildTierComparisons computes per-tier comparison data between left and right runs.
func buildTierComparisons(leftRuns, rightRuns []*store.Run) []TierComparison {
	leftTiers := aggregateByTier(leftRuns)
	rightTiers := aggregateByTier(rightRuns)

	// Collect all unique tiers, sorted.
	tierSet := make(map[int]bool)
	for t := range leftTiers {
		tierSet[t] = true
	}
	for t := range rightTiers {
		tierSet[t] = true
	}
	tiers := make([]int, 0, len(tierSet))
	for t := range tierSet {
		tiers = append(tiers, t)
	}
	sort.Ints(tiers)

	var result []TierComparison
	for _, tier := range tiers {
		tc := TierComparison{Tier: tier}

		leftTS := leftTiers[tier]
		rightTS := rightTiers[tier]

		var leftAvg, rightAvg float64
		if leftTS != nil && leftTS.totalCount > 0 {
			leftAvg = leftTS.totalCorr / float64(leftTS.totalCount)
			tc.LeftPassRate = fmt.Sprintf("%.1f%%", float64(leftTS.passCount)/float64(leftTS.totalCount)*100)
			tc.LeftAvgCorr = fmt.Sprintf("%.2f", leftAvg)
		} else {
			tc.LeftPassRate = "-"
			tc.LeftAvgCorr = "-"
		}
		if rightTS != nil && rightTS.totalCount > 0 {
			rightAvg = rightTS.totalCorr / float64(rightTS.totalCount)
			tc.RightPassRate = fmt.Sprintf("%.1f%%", float64(rightTS.passCount)/float64(rightTS.totalCount)*100)
			tc.RightAvgCorr = fmt.Sprintf("%.2f", rightAvg)
		} else {
			tc.RightPassRate = "-"
			tc.RightAvgCorr = "-"
		}

		if leftTS != nil && rightTS != nil {
			delta := formatDeltaFloat(rightAvg - leftAvg)
			// P1: Check CI overlap for per-tier pass rates.
			ll, lu := metrics.WilsonCI(leftTS.passCount, leftTS.totalCount)
			rl, ru := metrics.WilsonCI(rightTS.passCount, rightTS.totalCount)
			if !metrics.CIOverlaps(ll, lu, rl, ru) {
				tc.Significant = true
				delta += " *"
			}
			tc.DeltaCorrectness = delta
		} else {
			tc.DeltaCorrectness = "-"
		}

		result = append(result, tc)
	}
	return result
}

// aggregateByTier groups runs by tier and computes aggregate stats.
func aggregateByTier(runs []*store.Run) map[int]*tierRunStats {
	m := make(map[int]*tierRunStats)
	for _, r := range runs {
		ts, ok := m[r.Tier]
		if !ok {
			ts = &tierRunStats{}
			m[r.Tier] = ts
		}
		ts.totalCount++

		l1Pass := r.L1Build != nil && *r.L1Build
		l4AllPass := r.L4E2EPassed != nil && r.L4E2ETotal != nil &&
			*r.L4E2ETotal > 0 && *r.L4E2EPassed == *r.L4E2ETotal
		if l1Pass && l4AllPass {
			ts.passCount++
		}
		if r.CorrectnessScore != nil {
			ts.totalCorr += *r.CorrectnessScore
		}
	}
	return m
}
