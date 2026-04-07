package main

import (
	"fmt"
	"io"
	"os"
)

// runCmd executes the given subcommand with args, writing output to stdout.
// This function signature is part of the API contract and must not be changed.
func runCmd(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		fmt.Fprintln(stdout, "usage: taskcli <command>")
		fmt.Fprintln(stdout, "commands: list, add, help")
		return nil
	}

	switch args[0] {
	case "list":
		fmt.Fprintln(stdout, "No tasks found.")
		return nil
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("usage: taskcli add <title>")
		}
		fmt.Fprintf(stdout, "Added task: %s\n", args[1])
		return nil
	case "help":
		fmt.Fprintln(stdout, "usage: taskcli <command>")
		fmt.Fprintln(stdout, "commands: list, add, help")
		return nil
	// NOTE: "version" subcommand is missing - agent should add it
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func main() {
	if err := runCmd(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
