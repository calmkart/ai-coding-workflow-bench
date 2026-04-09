// Package report generates benchmark reports from stored results.
package report

import (
	_ "embed"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/calmkart/ai-coding-workflow-bench/internal/metrics"
	"github.com/calmkart/ai-coding-workflow-bench/internal/store"
)

//go:embed templates/summary.md.tmpl
var summaryTemplate string

// TaskDetail holds per-run data for the report.
type TaskDetail struct {
	ID          string
	Tier        int
	Type        string
	RunNumber   int
	L1          bool
	L2Pass      int
	L2Total     int
	L3Issues    int
	L4Pass      int
	L4Total     int
	Correctness float64
}

// TaskAggregate holds aggregated data for a task across multiple runs.
type TaskAggregate struct {
	ID             string
	Tier           int
	Type           string
	RunCount       int
	PassCount      int
	AvgCorrectness float64
	Runs           []TaskDetail
}

// TierSummary holds aggregated data for a tier across all its tasks.
type TierSummary struct {
	Tier           int
	TaskCount      int
	PassCount      int
	PassRate       string // "95.0%"
	AvgCorrectness string // "0.97"
}

// RubricSummary holds aggregated rubric scores for the report.
type RubricSummary struct {
	Correctness     float64
	Readability     float64
	Simplicity      float64
	Robustness      float64
	Minimality      float64
	Maintainability float64
	GoIdioms        float64
	Composite       float64
}

// TaskStability holds per-task stability data when RunsPerTask > 1.
type TaskStability struct {
	TaskID    string
	PassCount int
	TotalRuns int
	Stability string // "3/5 (60%)"
}

// SummaryData holds all data for the summary report template.
type SummaryData struct {
	Tag            string
	Workflow       string
	Date           string
	RunsPerTask    string // e.g. "3" or "1-3" if varies
	TotalTasks     int
	PassRate       float64
	PassRateCI     string // P1: "95.0% [87.2-98.6]"
	AvgCorrectness float64
	AvgWallTime    string // average execution time (e.g. "2m30s"), empty if no data
	AvgTokens      string // average total token count, empty if no data
	TotalCost      string // total estimated cost across all runs, empty if no token data
	Tasks          []TaskDetail
	TaskGroups     []TaskAggregate   // Non-empty when RunsPerTask > 1
	TierGroups     []TierSummary     // Per-tier aggregated summary
	StabilityData  []TaskStability   // Non-empty when RunsPerTask > 1
	MaxRunsPerTask int               // numeric max for template logic
	HasRubric            bool              // true if any run has rubric scores
	Rubric               RubricSummary     // average rubric scores across all scored runs
	ConsistencyWarnings  []string          // P13: boolean-to-score consistency warnings across all runs
	LowSampleWarning     string            // P1: non-empty if K < 5
}

// GenerateSummary renders a summary report for the given tag.
//
// @implements REQ-REPORT (generate summary report from stored runs)
func GenerateSummary(w io.Writer, runs []*store.Run) error {
	if len(runs) == 0 {
		return fmt.Errorf("no runs found")
	}

	tmpl, err := template.New("summary").Parse(summaryTemplate)
	if err != nil {
		return fmt.Errorf("parse summary template: %w", err)
	}

	data := buildSummaryData(runs)
	return tmpl.Execute(w, data)
}

func buildSummaryData(runs []*store.Run) SummaryData {
	data := SummaryData{
		Tag: runs[0].Tag,
	}

	// Determine workflow: collect distinct workflows from actual data.
	workflowSet := make(map[string]bool)
	for _, r := range runs {
		workflowSet[r.Workflow] = true
	}
	if len(workflowSet) == 1 {
		for wf := range workflowSet {
			data.Workflow = wf
		}
	} else {
		data.Workflow = "multiple"
	}

	// Determine date range from actual run timestamps.
	var minTime, maxTime time.Time
	for _, r := range runs {
		if minTime.IsZero() || r.StartedAt.Before(minTime) {
			minTime = r.StartedAt
		}
		if r.StartedAt.After(maxTime) {
			maxTime = r.StartedAt
		}
		if r.FinishedAt != nil {
			if r.FinishedAt.After(maxTime) {
				maxTime = *r.FinishedAt
			}
		}
	}
	minDateStr := minTime.Format("2006-01-02 15:04")
	maxDateStr := maxTime.Format("2006-01-02 15:04")
	if minDateStr == maxDateStr {
		data.Date = minDateStr
	} else if minTime.Format("2006-01-02") == maxTime.Format("2006-01-02") {
		// Same day, show time range
		data.Date = minTime.Format("2006-01-02 15:04") + " - " + maxTime.Format("15:04")
	} else {
		data.Date = minDateStr + " - " + maxDateStr
	}

	// Count runs per task to get RunsPerTask.
	taskRuns := make(map[string]int)
	for _, r := range runs {
		taskRuns[r.TaskID]++
	}
	data.TotalTasks = len(taskRuns)

	// Calculate min/max runs per task for display.
	var minRuns, maxRuns int
	if data.TotalTasks > 0 {
		first := true
		for _, count := range taskRuns {
			if first {
				minRuns = count
				maxRuns = count
				first = false
			} else {
				if count < minRuns {
					minRuns = count
				}
				if count > maxRuns {
					maxRuns = count
				}
			}
		}
	}
	data.MaxRunsPerTask = maxRuns
	if minRuns == maxRuns {
		data.RunsPerTask = fmt.Sprintf("%d", maxRuns)
	} else {
		data.RunsPerTask = fmt.Sprintf("%d-%d", minRuns, maxRuns)
	}

	var totalCorrectness float64
	var passCount int

	// Build per-run task details and track task ordering.
	taskOrder := []string{}
	taskSeen := make(map[string]bool)

	for _, r := range runs {
		td := TaskDetail{
			ID:        r.TaskID,
			Tier:      r.Tier,
			Type:      r.TaskType,
			RunNumber: r.RunNumber,
		}

		if r.L1Build != nil {
			td.L1 = *r.L1Build
		}
		if r.L2UtPassed != nil {
			td.L2Pass = *r.L2UtPassed
		}
		if r.L2UtTotal != nil {
			td.L2Total = *r.L2UtTotal
		}
		if r.L3LintIssues != nil {
			td.L3Issues = *r.L3LintIssues
		}
		if r.L4E2EPassed != nil {
			td.L4Pass = *r.L4E2EPassed
		}
		if r.L4E2ETotal != nil {
			td.L4Total = *r.L4E2ETotal
		}
		if r.CorrectnessScore != nil {
			td.Correctness = *r.CorrectnessScore
			totalCorrectness += *r.CorrectnessScore
		}

		// A run "passes" if L4 has all E2E tests passing.
		if td.L1 && td.L4Total > 0 && td.L4Pass == td.L4Total {
			passCount++
		}

		data.Tasks = append(data.Tasks, td)

		if !taskSeen[r.TaskID] {
			taskSeen[r.TaskID] = true
			taskOrder = append(taskOrder, r.TaskID)
		}
	}

	if len(runs) > 0 {
		data.PassRate = float64(passCount) / float64(len(runs)) * 100
		data.AvgCorrectness = totalCorrectness / float64(len(runs))
		// P1: Add Wilson Score CI for pass rate.
		data.PassRateCI = metrics.FormatCI(passCount, len(runs))
	}

	// P1: Low sample size warning.
	if metrics.IsLowSampleSize(len(runs)) {
		data.LowSampleWarning = "Low sample size (K < 5): results may not be statistically reliable"
	}

	// Calculate average wall time and token usage across runs.
	var totalWallSecs float64
	var wallTimeCount int
	var totalTokens int64
	var tokenCount int
	var totalCost float64
	for _, r := range runs {
		if r.WallTimeSecs != nil {
			totalWallSecs += *r.WallTimeSecs
			wallTimeCount++
		}
		if r.TotalTokens != nil {
			totalTokens += int64(*r.TotalTokens)
			tokenCount++
			// P8: Use pre-computed CostUSD from the run instead of recalculating
			// with hardcoded pricing.
			if r.CostUSD != nil {
				totalCost += *r.CostUSD
			}
		}
	}
	if wallTimeCount > 0 {
		avgSecs := totalWallSecs / float64(wallTimeCount)
		data.AvgWallTime = (time.Duration(avgSecs * float64(time.Second))).Truncate(time.Second).String()
	}
	if tokenCount > 0 {
		avgTok := totalTokens / int64(tokenCount)
		data.AvgTokens = fmt.Sprintf("%d", avgTok)
		data.TotalCost = fmt.Sprintf("$%.2f", totalCost)
	}

	// Calculate average rubric scores.
	var rubricCount int
	var rubricSum RubricSummary
	for _, r := range runs {
		if r.RubricComposite != nil {
			rubricCount++
			if r.RubricCorrectness != nil {
				rubricSum.Correctness += *r.RubricCorrectness
			}
			if r.RubricReadability != nil {
				rubricSum.Readability += *r.RubricReadability
			}
			if r.RubricSimplicity != nil {
				rubricSum.Simplicity += *r.RubricSimplicity
			}
			if r.RubricRobustness != nil {
				rubricSum.Robustness += *r.RubricRobustness
			}
			if r.RubricMinimality != nil {
				rubricSum.Minimality += *r.RubricMinimality
			}
			if r.RubricMaintainability != nil {
				rubricSum.Maintainability += *r.RubricMaintainability
			}
			if r.RubricGoIdioms != nil {
				rubricSum.GoIdioms += *r.RubricGoIdioms
			}
			rubricSum.Composite += *r.RubricComposite
		}
	}
	if rubricCount > 0 {
		data.HasRubric = true
		data.Rubric = RubricSummary{
			Correctness:     rubricSum.Correctness / float64(rubricCount),
			Readability:     rubricSum.Readability / float64(rubricCount),
			Simplicity:      rubricSum.Simplicity / float64(rubricCount),
			Robustness:      rubricSum.Robustness / float64(rubricCount),
			Minimality:      rubricSum.Minimality / float64(rubricCount),
			Maintainability: rubricSum.Maintainability / float64(rubricCount),
			GoIdioms:        rubricSum.GoIdioms / float64(rubricCount),
			Composite:       rubricSum.Composite / float64(rubricCount),
		}
	}

	// P13: Collect consistency warnings from all runs.
	warningSet := make(map[string]bool)
	for _, r := range runs {
		if r.RubricConsistencyWarnings != nil && *r.RubricConsistencyWarnings != "" {
			for _, w := range strings.Split(*r.RubricConsistencyWarnings, "; ") {
				warningSet[w] = true
			}
		}
	}
	if len(warningSet) > 0 {
		warnings := make([]string, 0, len(warningSet))
		for w := range warningSet {
			warnings = append(warnings, w)
		}
		sort.Strings(warnings)
		data.ConsistencyWarnings = warnings
	}

	// Build per-tier summary.
	tierMap := make(map[int]*TierSummary)
	for _, td := range data.Tasks {
		ts, ok := tierMap[td.Tier]
		if !ok {
			ts = &TierSummary{Tier: td.Tier}
			tierMap[td.Tier] = ts
		}
		ts.TaskCount++
		if td.L1 && td.L4Total > 0 && td.L4Pass == td.L4Total {
			ts.PassCount++
		}
	}
	// Collect tiers sorted by tier number.
	tierNums := make([]int, 0, len(tierMap))
	for t := range tierMap {
		tierNums = append(tierNums, t)
	}
	sort.Ints(tierNums)
	for _, tNum := range tierNums {
		ts := tierMap[tNum]
		// Compute per-tier average correctness.
		var tierCorr float64
		var tierCount int
		for _, td := range data.Tasks {
			if td.Tier == tNum {
				tierCorr += td.Correctness
				tierCount++
			}
		}
		if tierCount > 0 {
			ts.AvgCorrectness = fmt.Sprintf("%.2f", tierCorr/float64(tierCount))
		} else {
			ts.AvgCorrectness = "0.00"
		}
		if ts.TaskCount > 0 {
			ts.PassRate = fmt.Sprintf("%.1f%%", float64(ts.PassCount)/float64(ts.TaskCount)*100)
		} else {
			ts.PassRate = "0.0%"
		}
		data.TierGroups = append(data.TierGroups, *ts)
	}

	// Build per-task aggregates when there are multiple runs per task.
	if data.MaxRunsPerTask > 1 {
		// Group tasks by ID preserving order.
		taskRunsMap := make(map[string][]TaskDetail)
		for _, td := range data.Tasks {
			taskRunsMap[td.ID] = append(taskRunsMap[td.ID], td)
		}

		for _, taskID := range taskOrder {
			tds := taskRunsMap[taskID]
			agg := TaskAggregate{
				ID:       taskID,
				Tier:     tds[0].Tier,
				Type:     tds[0].Type,
				RunCount: len(tds),
				Runs:     tds,
			}
			var sumCorr float64
			taskPassCount := 0
			for _, td := range tds {
				sumCorr += td.Correctness
				if td.L1 && td.L4Total > 0 && td.L4Pass == td.L4Total {
					agg.PassCount++
					taskPassCount++
				}
			}
			if len(tds) > 0 {
				agg.AvgCorrectness = sumCorr / float64(len(tds))
			}
			data.TaskGroups = append(data.TaskGroups, agg)

			// Build stability entry.
			pct := float64(taskPassCount) / float64(len(tds)) * 100
			data.StabilityData = append(data.StabilityData, TaskStability{
				TaskID:    taskID,
				PassCount: taskPassCount,
				TotalRuns: len(tds),
				Stability: fmt.Sprintf("%d/%d (%.0f%%)", taskPassCount, len(tds), pct),
			})
		}
	}

	return data
}
