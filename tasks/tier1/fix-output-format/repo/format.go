package main

import (
	"fmt"
	"io"
)

// Task represents a task item.
type Task struct {
	ID     int
	Title  string
	Status string
}

// formatTasks writes tasks to w in a tabular format.
// This function signature is part of the API contract and must not be changed.
// BUG: uses simple \t which doesn't align columns properly.
func formatTasks(tasks []Task, w io.Writer) {
	// BUG: simple tab-separated output doesn't align when titles have different lengths
	fmt.Fprintf(w, "ID\tTitle\tStatus\n")
	for _, t := range tasks {
		fmt.Fprintf(w, "%d\t%s\t%s\n", t.ID, t.Title, t.Status)
	}
}
