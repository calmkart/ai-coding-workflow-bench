package engine

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
)

// TestProgressTracker_Report verifies that progressTracker.report produces correct output.
func TestProgressTracker_Report(t *testing.T) {
	pt := newProgressTracker(3)

	// Capture stdout.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	pt.report("tier1/task-a", 1, true)
	pt.report("tier1/task-b", 1, false)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// First report: [1/3] 33% done | pass: 100% | ETA: ...
	if !strings.Contains(output, "[1/3]") {
		t.Errorf("expected [1/3] in output, got:\n%s", output)
	}
	if !strings.Contains(output, "33%") {
		t.Errorf("expected 33%% in output, got:\n%s", output)
	}

	// Second report: [2/3] 67% done | pass: 50% | ETA: ...
	if !strings.Contains(output, "[2/3]") {
		t.Errorf("expected [2/3] in output, got:\n%s", output)
	}
	if !strings.Contains(output, "67%") {
		t.Errorf("expected 67%% in output, got:\n%s", output)
	}
}

// TestProgressTracker_PassRate verifies pass rate calculation.
func TestProgressTracker_PassRate(t *testing.T) {
	pt := newProgressTracker(4)

	pt.report("task-1", 1, true)
	pt.report("task-2", 1, true)
	pt.report("task-3", 1, false)

	completed := pt.completed.Load()
	passed := pt.passed.Load()

	if completed != 3 {
		t.Errorf("expected 3 completed, got %d", completed)
	}
	if passed != 2 {
		t.Errorf("expected 2 passed, got %d", passed)
	}

	// Pass rate should be 66.67%.
	rate := float64(passed) / float64(completed) * 100
	if rate < 66 || rate > 67 {
		t.Errorf("expected pass rate ~67%%, got %.1f%%", rate)
	}
}

// TestProgressTracker_ConcurrentReport verifies thread safety of report.
func TestProgressTracker_ConcurrentReport(t *testing.T) {
	pt := newProgressTracker(100)

	// Redirect stdout to avoid cluttering test output.
	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			pt.report(fmt.Sprintf("task-%d", n), 1, n%2 == 0)
		}(i)
	}
	wg.Wait()

	w.Close()
	os.Stdout = old

	if pt.completed.Load() != 100 {
		t.Errorf("expected 100 completed, got %d", pt.completed.Load())
	}
	if pt.passed.Load() != 50 {
		t.Errorf("expected 50 passed, got %d", pt.passed.Load())
	}
}

// TestProgressTracker_ETADecreasesToZero verifies that ETA shows "0s" when all runs are done.
func TestProgressTracker_ETADecreasesToZero(t *testing.T) {
	pt := newProgressTracker(1)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	pt.report("task-1", 1, true)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// When all runs are done, ETA should be "0s".
	if !strings.Contains(output, "ETA: 0s") {
		t.Errorf("expected 'ETA: 0s' when all runs complete, got:\n%s", output)
	}
	if !strings.Contains(output, "100%") {
		t.Errorf("expected '100%%' when all runs complete, got:\n%s", output)
	}
}
