package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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

// SMELL: These are standalone functions, not implementing any interface.
// Should be refactored to Command implementations.

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

func cmdDone(idStr string) {
	loadTasks()
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid id")
		os.Exit(1)
	}
	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Done = true
			saveTasks()
			fmt.Printf("Task %d marked as done\n", id)
			return
		}
	}
	fmt.Fprintf(os.Stderr, "task %d not found\n", id)
	os.Exit(1)
}

func cmdDelete(idStr string) {
	loadTasks()
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid id")
		os.Exit(1)
	}
	for i := range tasks {
		if tasks[i].ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			saveTasks()
			fmt.Printf("Task %d deleted\n", id)
			return
		}
	}
	fmt.Fprintf(os.Stderr, "task %d not found\n", id)
	os.Exit(1)
}
