// Package report generates benchmark reports from stored results.
package report

import (
	_ "embed"
	"fmt"
	htmltemplate "html/template"
	"io"
	"math"
	"strings"
	"time"

	"github.com/calmkart/ai-coding-workflow-bench/internal/store"
)

// TrendTagData holds aggregated metrics for a single tag in a trend report.
type TrendTagData struct {
	Tag             string
	Date            string
	PassRate        float64 // 0-100
	AvgCorrectness  float64 // 0-1
	AvgWallTime     string  // formatted duration
	AvgWallTimeSecs float64 // for delta computation
	Tasks           int
}

// TrendData holds all data for the trend report.
type TrendData struct {
	Tags []TrendTagData

	// Deltas between first and last tag.
	DeltaPassRate    string // e.g. "+15.0%"
	DeltaCorrectness string // e.g. "+0.13"
	DeltaWallTime    string // e.g. "-26%"
}

// TagRunSet pairs a tag name with its runs, preserving caller-specified order.
type TagRunSet struct {
	Tag  string
	Runs []*store.Run
}

// BuildTrendData computes TrendData from runs grouped by tag.
// Tags are ordered as provided by the caller.
//
// @implements P23 (trend report data computation)
func BuildTrendData(tagRunSets []TagRunSet) TrendData {
	var data TrendData

	for _, trs := range tagRunSets {
		runs := trs.Runs
		if len(runs) == 0 {
			continue
		}

		td := TrendTagData{
			Tag:   trs.Tag,
			Tasks: countDistinctTasks(runs),
		}

		// Compute date from earliest run.
		var earliest time.Time
		for _, r := range runs {
			if earliest.IsZero() || r.StartedAt.Before(earliest) {
				earliest = r.StartedAt
			}
		}
		td.Date = earliest.Format("2006-01-02")

		// Compute pass rate: a run "passes" if L1=true and L4 all pass.
		passCount := 0
		for _, r := range runs {
			if r.L1Build != nil && *r.L1Build &&
				r.L4E2EPassed != nil && r.L4E2ETotal != nil &&
				*r.L4E2ETotal > 0 && *r.L4E2EPassed == *r.L4E2ETotal {
				passCount++
			}
		}
		if len(runs) > 0 {
			td.PassRate = float64(passCount) / float64(len(runs)) * 100
		}

		// Compute average correctness.
		var totalCorr float64
		var corrCount int
		for _, r := range runs {
			if r.CorrectnessScore != nil {
				totalCorr += *r.CorrectnessScore
				corrCount++
			}
		}
		if corrCount > 0 {
			td.AvgCorrectness = totalCorr / float64(corrCount)
		}

		// Compute average wall time.
		var totalWall float64
		var wallCount int
		for _, r := range runs {
			if r.WallTimeSecs != nil {
				totalWall += *r.WallTimeSecs
				wallCount++
			}
		}
		if wallCount > 0 {
			avgSecs := totalWall / float64(wallCount)
			td.AvgWallTimeSecs = avgSecs
			td.AvgWallTime = (time.Duration(avgSecs * float64(time.Second))).Truncate(time.Second).String()
		} else {
			td.AvgWallTime = "-"
		}

		data.Tags = append(data.Tags, td)
	}

	// Compute deltas between first and last tag.
	if len(data.Tags) >= 2 {
		first := data.Tags[0]
		last := data.Tags[len(data.Tags)-1]

		// Pass rate delta.
		deltaRate := last.PassRate - first.PassRate
		if deltaRate >= 0 {
			data.DeltaPassRate = fmt.Sprintf("+%.1f%%", deltaRate)
		} else {
			data.DeltaPassRate = fmt.Sprintf("%.1f%%", deltaRate)
		}

		// Correctness delta.
		deltaCorr := last.AvgCorrectness - first.AvgCorrectness
		if deltaCorr >= 0 {
			data.DeltaCorrectness = fmt.Sprintf("+%.2f", deltaCorr)
		} else {
			data.DeltaCorrectness = fmt.Sprintf("%.2f", deltaCorr)
		}

		// Wall time delta (percentage).
		if first.AvgWallTimeSecs > 0 && last.AvgWallTimeSecs > 0 {
			deltaWallPct := (last.AvgWallTimeSecs - first.AvgWallTimeSecs) / first.AvgWallTimeSecs * 100
			deltaWallPct = math.Round(deltaWallPct)
			if deltaWallPct >= 0 {
				data.DeltaWallTime = fmt.Sprintf("+%.0f%%", deltaWallPct)
			} else {
				data.DeltaWallTime = fmt.Sprintf("%.0f%%", deltaWallPct)
			}
		} else {
			data.DeltaWallTime = "-"
		}
	}

	return data
}

// countDistinctTasks counts unique task IDs in runs.
func countDistinctTasks(runs []*store.Run) int {
	seen := make(map[string]bool)
	for _, r := range runs {
		seen[r.TaskID] = true
	}
	return len(seen)
}

// GenerateTrend renders a trend report in markdown format.
//
// @implements P23 (trend report generation - markdown)
func GenerateTrend(w io.Writer, tagRunSets []TagRunSet) error {
	if len(tagRunSets) == 0 {
		return fmt.Errorf("no tags provided for trend report")
	}

	data := BuildTrendData(tagRunSets)
	if len(data.Tags) == 0 {
		return fmt.Errorf("no data found for any tag")
	}

	return renderTrendMarkdown(w, data)
}

// renderTrendMarkdown writes the trend report in markdown format.
func renderTrendMarkdown(w io.Writer, data TrendData) error {
	var sb strings.Builder

	sb.WriteString("# Trend Report\n\n")
	sb.WriteString("| Tag | Date | Pass Rate | Avg Correctness | Avg Wall Time | Tasks |\n")
	sb.WriteString("|-----|------|-----------|-----------------|---------------|-------|\n")

	for _, t := range data.Tags {
		sb.WriteString(fmt.Sprintf("| %s | %s | %.1f%% | %.2f | %s | %d |\n",
			t.Tag, t.Date, t.PassRate, t.AvgCorrectness, t.AvgWallTime, t.Tasks))
	}

	// Trend summary line.
	if len(data.Tags) >= 2 {
		sb.WriteString(fmt.Sprintf("\nTrend: Pass Rate %s, Correctness %s, Wall Time %s\n",
			data.DeltaPassRate, data.DeltaCorrectness, data.DeltaWallTime))
	}

	_, err := io.WriteString(w, sb.String())
	return err
}

//go:embed templates/trend.html.tmpl
var trendHTMLTemplate string

// GenerateTrendHTML renders a trend report in HTML format.
//
// @implements P23 (trend report generation - HTML)
func GenerateTrendHTML(w io.Writer, tagRunSets []TagRunSet) error {
	if len(tagRunSets) == 0 {
		return fmt.Errorf("no tags provided for trend report")
	}

	data := BuildTrendData(tagRunSets)
	if len(data.Tags) == 0 {
		return fmt.Errorf("no data found for any tag")
	}

	tmpl, err := htmltemplate.New("trend_html").Funcs(htmltemplate.FuncMap{
		"deltaClass": deltaClass,
	}).Parse(trendHTMLTemplate)
	if err != nil {
		return fmt.Errorf("parse trend HTML template: %w", err)
	}

	return tmpl.Execute(w, data)
}
