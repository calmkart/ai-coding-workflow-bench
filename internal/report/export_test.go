package report

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/calmkart/ai-coding-workflow-bench/internal/store"
)

func TestExportJSON_Basic(t *testing.T) {
	runs := []*store.Run{
		makeRun("r1", "test-tag", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 60),
		makeRun("r2", "test-tag", "tier1/task-b", false, 2, 8, 0, 5, 0.3, 120),
	}

	var buf bytes.Buffer
	err := ExportJSON(&buf, runs)
	if err != nil {
		t.Fatalf("ExportJSON: %v", err)
	}

	// Parse the JSON output.
	var records []ExportRecord
	if err := json.Unmarshal(buf.Bytes(), &records); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if records[0].RunID != "r1" {
		t.Errorf("expected run_id=r1, got %q", records[0].RunID)
	}
	if records[0].Tag != "test-tag" {
		t.Errorf("expected tag=test-tag, got %q", records[0].Tag)
	}
	if records[0].TaskID != "tier1/task-a" {
		t.Errorf("expected task_id=tier1/task-a, got %q", records[0].TaskID)
	}
	if records[0].Correctness == nil || *records[0].Correctness != 1.0 {
		t.Errorf("expected correctness=1.0, got %v", records[0].Correctness)
	}
	if records[1].L1Build == nil || *records[1].L1Build != false {
		t.Errorf("expected l1_build=false for r2")
	}
}

func TestExportJSON_Empty(t *testing.T) {
	var buf bytes.Buffer
	err := ExportJSON(&buf, nil)
	if err != nil {
		t.Fatalf("ExportJSON empty: %v", err)
	}
	output := strings.TrimSpace(buf.String())
	if output != "[]" {
		t.Errorf("expected empty JSON array, got %q", output)
	}
}

func TestExportCSV_Basic(t *testing.T) {
	runs := []*store.Run{
		makeRun("r1", "test-tag", "tier1/task-a", true, 8, 8, 5, 5, 1.0, 60),
	}

	var buf bytes.Buffer
	err := ExportCSV(&buf, runs)
	if err != nil {
		t.Fatalf("ExportCSV: %v", err)
	}

	r := csv.NewReader(strings.NewReader(buf.String()))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("read CSV: %v", err)
	}

	// Header + 1 data row.
	if len(records) != 2 {
		t.Fatalf("expected 2 CSV rows (header+data), got %d", len(records))
	}

	// Check header.
	header := records[0]
	if header[0] != "run_id" {
		t.Errorf("expected first header column 'run_id', got %q", header[0])
	}
	if header[3] != "task_id" {
		t.Errorf("expected header[3]='task_id', got %q", header[3])
	}

	// Check data row.
	data := records[1]
	if data[0] != "r1" {
		t.Errorf("expected run_id=r1, got %q", data[0])
	}
	if data[1] != "test-tag" {
		t.Errorf("expected tag=test-tag, got %q", data[1])
	}
}

func TestExportCSV_Empty(t *testing.T) {
	var buf bytes.Buffer
	err := ExportCSV(&buf, nil)
	if err != nil {
		t.Fatalf("ExportCSV empty: %v", err)
	}

	r := csv.NewReader(strings.NewReader(buf.String()))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("read CSV: %v", err)
	}

	// Should have header only.
	if len(records) != 1 {
		t.Fatalf("expected 1 CSV row (header only), got %d", len(records))
	}
}

func TestExportCSV_NilFields(t *testing.T) {
	now := time.Now()
	finished := now.Add(1 * time.Minute)
	runs := []*store.Run{
		{
			ID: "r-nil", Tag: "test", Workflow: "vanilla",
			TaskID: "tier1/task", Tier: 1, TaskType: "http-server",
			RunNumber: 1, Status: "completed",
			StartedAt: now, FinishedAt: &finished,
			// All optional fields left nil.
		},
	}

	var buf bytes.Buffer
	err := ExportCSV(&buf, runs)
	if err != nil {
		t.Fatalf("ExportCSV nil fields: %v", err)
	}

	r := csv.NewReader(strings.NewReader(buf.String()))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("read CSV: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(records))
	}

	// Nil fields should be empty strings.
	data := records[1]
	// l1_build (index 8) should be empty.
	if data[8] != "" {
		t.Errorf("expected empty l1_build for nil, got %q", data[8])
	}
}
