package main

import (
	"fmt"
	"os"
)

// SMELL: Giant switch-case in main. Every new command requires modifying this function.
// Should be refactored to Command interface + Registry pattern.
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: taskcli <command> [args]")
		os.Exit(1)
	}

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
	case "delete":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: taskcli delete <id>")
			os.Exit(1)
		}
		cmdDelete(os.Args[2])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
