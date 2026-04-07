package main

import "testing"

func TestTaskStruct(t *testing.T) {
	task := Task{ID: 1, Title: "test", Done: false}
	if task.Title != "test" {
		t.Fatalf("expected 'test', got '%s'", task.Title)
	}
}
