package engine

import (
	"testing"

	"github.com/calmkart/ai-coding-workflow-bench/internal/config"
)

func TestShardTasks(t *testing.T) {
	// Create 8 tasks.
	tasks := make([]*config.TaskMeta, 8)
	for i := range tasks {
		tasks[i] = &config.TaskMeta{ID: string(rune('A' + i))}
	}

	tests := []struct {
		name     string
		index    int
		total    int
		wantIDs  []string
	}{
		{
			name:    "shard 1/4",
			index:   1,
			total:   4,
			wantIDs: []string{"A", "E"}, // indices 0, 4
		},
		{
			name:    "shard 2/4",
			index:   2,
			total:   4,
			wantIDs: []string{"B", "F"}, // indices 1, 5
		},
		{
			name:    "shard 3/4",
			index:   3,
			total:   4,
			wantIDs: []string{"C", "G"}, // indices 2, 6
		},
		{
			name:    "shard 4/4",
			index:   4,
			total:   4,
			wantIDs: []string{"D", "H"}, // indices 3, 7
		},
		{
			name:    "shard 1/1 returns all",
			index:   1,
			total:   1,
			wantIDs: []string{"A", "B", "C", "D", "E", "F", "G", "H"},
		},
		{
			name:    "shard 1/2",
			index:   1,
			total:   2,
			wantIDs: []string{"A", "C", "E", "G"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShardTasks(tasks, tt.index, tt.total)
			if len(got) != len(tt.wantIDs) {
				t.Fatalf("ShardTasks(%d, %d) returned %d tasks, want %d", tt.index, tt.total, len(got), len(tt.wantIDs))
			}
			for i, want := range tt.wantIDs {
				if got[i].ID != want {
					t.Errorf("ShardTasks(%d, %d)[%d].ID = %q, want %q", tt.index, tt.total, i, got[i].ID, want)
				}
			}
		})
	}
}

func TestShardTasks_InvalidArgs(t *testing.T) {
	tasks := []*config.TaskMeta{{ID: "A"}, {ID: "B"}}

	// Invalid shard params should return all tasks.
	tests := []struct {
		name  string
		index int
		total int
	}{
		{"zero total", 1, 0},
		{"zero index", 0, 4},
		{"index > total", 5, 4},
		{"negative index", -1, 4},
		{"negative total", 1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShardTasks(tasks, tt.index, tt.total)
			if len(got) != len(tasks) {
				t.Errorf("ShardTasks(%d, %d) returned %d tasks, want %d (all)", tt.index, tt.total, len(got), len(tasks))
			}
		})
	}
}

func TestShardTasks_EmptyTasks(t *testing.T) {
	got := ShardTasks(nil, 1, 4)
	if got != nil {
		t.Errorf("expected nil for empty tasks, got %v", got)
	}
}

func TestShardTasks_CoverageComplete(t *testing.T) {
	// Verify all tasks are covered exactly once across all shards.
	tasks := make([]*config.TaskMeta, 10)
	for i := range tasks {
		tasks[i] = &config.TaskMeta{ID: string(rune('0' + i))}
	}

	total := 3
	seen := make(map[string]int)
	for shard := 1; shard <= total; shard++ {
		got := ShardTasks(tasks, shard, total)
		for _, t := range got {
			seen[t.ID]++
		}
	}

	// Every task should appear exactly once.
	for i := 0; i < 10; i++ {
		id := string(rune('0' + i))
		if count, ok := seen[id]; !ok || count != 1 {
			t.Errorf("task %q appeared %d times (want exactly 1)", id, seen[id])
		}
	}
}
