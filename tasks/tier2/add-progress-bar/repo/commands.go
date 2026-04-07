package main

import (
	"encoding/json"
	"fmt"
	"os"
	"io"
	"strings"
)

type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

var tasks []Task
var nextID int = 1

const dataFile = "tasks.json"

func loadTasks() {
	data, err := os.ReadFile(dataFile)
	if err != nil {
		return
	}
	json.Unmarshal(data, &tasks)
	for _, t := range tasks {
		if t.ID >= nextID {
			nextID = t.ID + 1
		}
	}
}

func saveTasks() {
	data, _ := json.MarshalIndent(tasks, "", "  ")
	os.WriteFile(dataFile, data, 0644)
}

func cmdAdd(title string) {
	loadTasks()
	task := Task{ID: nextID, Title: title, Done: false}
	nextID++
	tasks = append(tasks, task)
	saveTasks()
	fmt.Printf("Added task %d: %s\n", task.ID, task.Title)
}

func cmdList() {
	loadTasks()
	if len(tasks) == 0 {
		fmt.Println("No tasks.")
		return
	}
	for _, t := range tasks {
		status := " "
		if t.Done {
			status = "x"
		}
		fmt.Printf("[%s] %d: %s\n", status, t.ID, t.Title)
	}
}

// ProgressBar displays a text-based progress bar.
// TODO: Implement properly with Update method.
type ProgressBar struct {
	Total   int
	Width   int
	Writer  io.Writer
}

// NewProgressBar creates a new progress bar.
func NewProgressBar(total int, w io.Writer) *ProgressBar {
	return &ProgressBar{Total: total, Width: 20, Writer: w}
}

// Update displays the current progress.
// BUG: Stub - just prints a number instead of a proper bar.
func (p *ProgressBar) Update(current int) {
	fmt.Fprintf(p.Writer, "%d/%d\n", current, p.Total)
}

func cmdProcess() {
	loadTasks()
	if len(tasks) == 0 {
		fmt.Println("No tasks to process.")
		return
	}

	bar := NewProgressBar(len(tasks), os.Stdout)
	for i, t := range tasks {
		_ = t
		_ = strings.TrimSpace // just to use strings
		bar.Update(i + 1)
	}
	fmt.Println("Processing complete!")
}
