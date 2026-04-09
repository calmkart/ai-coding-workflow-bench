// Package report generates benchmark reports from stored results.
package report

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/calmkart/ai-coding-workflow-bench/internal/store"
)

//go:embed templates/summary.html.tmpl
var summaryHTMLTemplate string

//go:embed templates/compare.html.tmpl
var compareHTMLTemplate string

// htmlFuncMap provides template helper functions for HTML reports.
var htmlFuncMap = template.FuncMap{
	"sub": func(a, b float64) float64 { return a - b },
	// rubricPct converts a 0-5 rubric score to a 0-100 percentage for CSS bar widths.
	"rubricPct": func(score float64) float64 { return score / 5.0 * 100.0 },
	// passRateClass returns a CSS class based on the pass rate string.
	// Green (>80%), yellow (50-80%), red (<50%).
	"passRateClass": passRateClass,
	// deltaClass returns a CSS class for a delta string ("+X" -> positive, "-X" -> negative).
	"deltaClass": deltaClass,
	// deltaClassFloat returns a CSS class for a numeric delta value.
	"deltaClassFloat": deltaClassFloat,
	// passIcon renders a pass/fail indicator as a styled HTML span.
	"passIcon": passIcon,
}

// GenerateSummaryHTML renders a summary report as a single-file HTML document.
//
// The output includes inline CSS with no external dependencies. It contains:
//   - Overall summary table
//   - Per-Tier summary with color-coded pass rates (green >80%, yellow 50-80%, red <50%)
//   - Per-Task detail table (L1=FAIL highlighted in red)
//   - CSS bar chart for rubric dimensions (when rubric data is present)
//
// @implements P14 (HTML summary report generation)
func GenerateSummaryHTML(w io.Writer, runs []*store.Run) error {
	if len(runs) == 0 {
		return fmt.Errorf("no runs found")
	}

	tmpl, err := template.New("summary_html").Funcs(htmlFuncMap).Parse(summaryHTMLTemplate)
	if err != nil {
		return fmt.Errorf("parse summary HTML template: %w", err)
	}

	data := buildSummaryData(runs)
	return tmpl.Execute(w, data)
}

// GenerateComparisonHTML renders a comparison report as a single-file HTML document.
//
// The output includes inline CSS with no external dependencies. It contains:
//   - Side-by-side overall metrics with color-coded deltas
//   - Per-Tier comparison with pass rate colors
//   - Per-Task comparison with pass/fail icons
//
// @implements P14 (HTML comparison report generation)
func GenerateComparisonHTML(w io.Writer, leftRuns, rightRuns []*store.Run, leftTag, rightTag string) error {
	if len(leftRuns) == 0 {
		return fmt.Errorf("no runs found for left tag %q", leftTag)
	}
	if len(rightRuns) == 0 {
		return fmt.Errorf("no runs found for right tag %q", rightTag)
	}

	tmpl, err := template.New("compare_html").Funcs(htmlFuncMap).Parse(compareHTMLTemplate)
	if err != nil {
		return fmt.Errorf("parse compare HTML template: %w", err)
	}

	data := buildCompareData(leftRuns, rightRuns, leftTag, rightTag)
	return tmpl.Execute(w, data)
}

// passRateClass returns a CSS class name based on a pass rate string like "95.0%".
// Returns "rate-green" for >80%, "rate-yellow" for 50-80%, "rate-red" for <50%.
func passRateClass(rateStr string) string {
	rateStr = strings.TrimSuffix(rateStr, "%")
	rateStr = strings.TrimSpace(rateStr)
	if rateStr == "-" || rateStr == "" {
		return ""
	}
	var rate float64
	if _, err := fmt.Sscanf(rateStr, "%f", &rate); err != nil {
		return ""
	}
	switch {
	case rate > 80:
		return "rate-green"
	case rate >= 50:
		return "rate-yellow"
	default:
		return "rate-red"
	}
}

// deltaClass returns a CSS class for a delta string.
// Strings starting with "+" get "delta-pos", "-" get "delta-neg", otherwise "delta-zero".
func deltaClass(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return "delta-zero"
	}
	if strings.HasPrefix(s, "+0.0%") || s == "+0.00" {
		return "delta-zero"
	}
	if strings.HasPrefix(s, "+") {
		return "delta-pos"
	}
	if strings.HasPrefix(s, "-") {
		return "delta-neg"
	}
	return "delta-zero"
}

// deltaClassFloat returns a CSS class for a numeric delta value.
func deltaClassFloat(d float64) string {
	const epsilon = 1e-9
	if d > epsilon {
		return "delta-pos"
	}
	if d < -epsilon {
		return "delta-neg"
	}
	return "delta-zero"
}

// passIcon renders a pass/fail string as an HTML span with appropriate styling.
func passIcon(s string) template.HTML {
	switch s {
	case "\u2705":
		return template.HTML(`<span class="pass">&#x2705;</span>`)
	case "\u274c":
		return template.HTML(`<span class="fail">&#x274C;</span>`)
	default:
		return template.HTML(template.HTMLEscapeString(s))
	}
}
