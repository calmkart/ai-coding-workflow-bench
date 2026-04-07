package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormatTasksNotEmpty(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "Test", Status: "done"},
	}
	var buf bytes.Buffer
	formatTasks(tasks, &buf)

	if buf.Len() == 0 {
		t.Fatal("expected non-empty output")
	}
}

func TestFormatTasksContainsTitle(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "Buy milk", Status: "pending"},
	}
	var buf bytes.Buffer
	formatTasks(tasks, &buf)

	if !strings.Contains(buf.String(), "Buy milk") {
		t.Fatal("output should contain task title")
	}
}

func TestFormatTasksEmpty(t *testing.T) {
	var buf bytes.Buffer
	formatTasks([]Task{}, &buf)

	// Should at least have header
	if !strings.Contains(buf.String(), "ID") {
		t.Fatal("empty task list should still have header")
	}
}
