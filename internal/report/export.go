// Package report generates benchmark reports from stored results.
package report

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/calmkart/ai-coding-workflow-bench/internal/store"
)

// ExportRecord represents a single run in the export output (JSON/CSV).
type ExportRecord struct {
	RunID            string   `json:"run_id"`
	Tag              string   `json:"tag"`
	Workflow         string   `json:"workflow"`
	TaskID           string   `json:"task_id"`
	Tier             int      `json:"tier"`
	TaskType         string   `json:"task_type"`
	RunNumber        int      `json:"run_number"`
	Status           string   `json:"status"`
	L1Build          *bool    `json:"l1_build"`
	L2Passed         *int     `json:"l2_passed"`
	L2Total          *int     `json:"l2_total"`
	L3Issues         *int     `json:"l3_issues"`
	L4Passed         *int     `json:"l4_passed"`
	L4Total          *int     `json:"l4_total"`
	Correctness      *float64 `json:"correctness"`
	EfficiencyScore  *float64 `json:"efficiency_score,omitempty"`
	FinalScore       *float64 `json:"final_score,omitempty"`
	WallTimeSecs     *float64 `json:"wall_time_secs,omitempty"`
	InputTokens      *int     `json:"input_tokens,omitempty"`
	OutputTokens     *int     `json:"output_tokens,omitempty"`
	TotalTokens      *int     `json:"total_tokens,omitempty"`
	CostUSD          *float64 `json:"cost_usd,omitempty"`
	RubricComposite  *float64 `json:"rubric_composite,omitempty"`
}

// ExportJSON writes runs as a JSON array to w.
//
// @implements P3 (data export - JSON format)
func ExportJSON(w io.Writer, runs []*store.Run) error {
	records := make([]ExportRecord, 0, len(runs))
	for _, r := range runs {
		records = append(records, toExportRecord(r))
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(records)
}

// ExportCSV writes runs as CSV (header + rows) to w.
//
// @implements P3 (data export - CSV format)
func ExportCSV(w io.Writer, runs []*store.Run) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	// Write header.
	header := []string{
		"run_id", "tag", "workflow", "task_id", "tier", "task_type",
		"run_number", "status", "l1_build", "l2_passed", "l2_total",
		"l3_issues", "l4_passed", "l4_total", "correctness",
		"efficiency_score", "final_score", "wall_time_secs",
		"input_tokens", "output_tokens", "total_tokens",
		"cost_usd", "rubric_composite",
	}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}

	// Write rows.
	for _, r := range runs {
		rec := toExportRecord(r)
		row := []string{
			rec.RunID, rec.Tag, rec.Workflow, rec.TaskID,
			strconv.Itoa(rec.Tier), rec.TaskType,
			strconv.Itoa(rec.RunNumber), rec.Status,
			optBool(rec.L1Build), optInt(rec.L2Passed), optInt(rec.L2Total),
			optInt(rec.L3Issues), optInt(rec.L4Passed), optInt(rec.L4Total),
			optFloat(rec.Correctness), optFloat(rec.EfficiencyScore),
			optFloat(rec.FinalScore), optFloat(rec.WallTimeSecs),
			optInt(rec.InputTokens), optInt(rec.OutputTokens),
			optInt(rec.TotalTokens), optFloat(rec.CostUSD),
			optFloat(rec.RubricComposite),
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("write CSV row: %w", err)
		}
	}
	return nil
}

func toExportRecord(r *store.Run) ExportRecord {
	return ExportRecord{
		RunID:           r.ID,
		Tag:             r.Tag,
		Workflow:        r.Workflow,
		TaskID:          r.TaskID,
		Tier:            r.Tier,
		TaskType:        r.TaskType,
		RunNumber:       r.RunNumber,
		Status:          r.Status,
		L1Build:         r.L1Build,
		L2Passed:        r.L2UtPassed,
		L2Total:         r.L2UtTotal,
		L3Issues:        r.L3LintIssues,
		L4Passed:        r.L4E2EPassed,
		L4Total:         r.L4E2ETotal,
		Correctness:     r.CorrectnessScore,
		EfficiencyScore: r.EfficiencyScore,
		FinalScore:      r.FinalScore,
		WallTimeSecs:    r.WallTimeSecs,
		InputTokens:     r.InputTokens,
		OutputTokens:    r.OutputTokens,
		TotalTokens:     r.TotalTokens,
		CostUSD:         r.CostUSD,
		RubricComposite: r.RubricComposite,
	}
}

func optBool(b *bool) string {
	if b == nil {
		return ""
	}
	if *b {
		return "true"
	}
	return "false"
}

func optInt(i *int) string {
	if i == nil {
		return ""
	}
	return strconv.Itoa(*i)
}

func optFloat(f *float64) string {
	if f == nil {
		return ""
	}
	return strconv.FormatFloat(*f, 'f', 4, 64)
}
