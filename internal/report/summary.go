// Package report generates benchmark reports from stored results.
package report

import (
	_ "embed"
	"fmt"
	"io"
	"text/template"
	"time"

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

// SummaryData holds all data for the summary report template.
type SummaryData struct {
	Tag            string
	Workflow       string
	Date           string
	RunsPerTask    string // e.g. "3" or "1-3" if varies
	TotalTasks     int
	PassRate       float64
	AvgCorrectness float64
	Tasks          []TaskDetail
	TaskGroups     []TaskAggregate // Non-empty when RunsPerTask > 1
	MaxRunsPerTask int             // numeric max for template logic
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
			for _, td := range tds {
				sumCorr += td.Correctness
				if td.L1 && td.L4Total > 0 && td.L4Pass == td.L4Total {
					agg.PassCount++
				}
			}
			if len(tds) > 0 {
				agg.AvgCorrectness = sumCorr / float64(len(tds))
			}
			data.TaskGroups = append(data.TaskGroups, agg)
		}
	}

	return data
}
