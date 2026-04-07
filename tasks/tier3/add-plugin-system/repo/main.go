package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: taskcli <command> [args]")
		os.Exit(1)
	}

	// PROBLEM: All commands hardcoded. Cannot extend without modifying source.
	switch os.Args[1] {
	case "add":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: taskcli add <title>")
			os.Exit(1)
		}
		cmdAdd(os.Args[2])
	case "list":
		cmdList()
	case "done":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: taskcli done <id>")
			os.Exit(1)
		}
		cmdDone(os.Args[2])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
