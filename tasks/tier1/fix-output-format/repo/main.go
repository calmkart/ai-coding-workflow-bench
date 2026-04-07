package main

import (
	"fmt"
	"os"
)

func main() {
	tasks := []Task{
		{ID: 1, Title: "Buy groceries", Status: "done"},
		{ID: 2, Title: "Write documentation for the API", Status: "pending"},
		{ID: 3, Title: "Fix bug", Status: "in-progress"},
	}

	formatTasks(tasks, os.Stdout)
	fmt.Println()
}
