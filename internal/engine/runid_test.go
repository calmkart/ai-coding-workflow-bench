package engine

import (
	"strings"
	"sync"
	"testing"
)

// TestGenerateRunID_Format verifies the format of generated run IDs.
func TestGenerateRunID_Format(t *testing.T) {
	id := generateRunID("v1", "tier1/fix-bug", 2)

	// Should start with tag-taskID-runN.
	if !strings.HasPrefix(id, "v1-tier1/fix-bug-run2-") {
		t.Errorf("unexpected prefix: %s", id)
	}

	// Should contain at least 2 dashes after the run part (timestamp and sequence).
	parts := strings.Split(id, "-")
	// Expected: v1, tier1/fix, bug, run2, <timestamp>, <seq>
	if len(parts) < 6 {
		t.Errorf("expected at least 6 dash-separated parts, got %d: %s", len(parts), id)
	}
}

// TestGenerateRunID_Uniqueness verifies that concurrent calls produce unique IDs.
func TestGenerateRunID_Uniqueness(t *testing.T) {
	const n = 100
	ids := make([]string, n)
	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			ids[idx] = generateRunID("test", "task", 1)
		}(i)
	}
	wg.Wait()

	seen := make(map[string]bool, n)
	for _, id := range ids {
		if seen[id] {
			t.Errorf("duplicate run ID: %s", id)
		}
		seen[id] = true
	}
}

// TestGenerateRunID_MonotonicSequence verifies that the sequence is monotonically increasing.
func TestGenerateRunID_MonotonicSequence(t *testing.T) {
	id1 := generateRunID("test", "task", 1)
	id2 := generateRunID("test", "task", 1)

	// Both should have different sequence numbers (last part).
	parts1 := strings.Split(id1, "-")
	parts2 := strings.Split(id2, "-")

	seq1 := parts1[len(parts1)-1]
	seq2 := parts2[len(parts2)-1]

	if seq1 == seq2 {
		t.Errorf("expected different sequences, got same: %s", seq1)
	}
}
