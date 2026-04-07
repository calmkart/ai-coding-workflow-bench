package main

import (
	"fmt"
	"io"
	"os"
)

// runApp executes the CLI and returns the exit code.
// This function signature is part of the API contract and must not be changed.
func runApp(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "error: no command specified")
		// BUG: returns 0 on error, should return 1
		return 0
	}

	switch args[0] {
	case "list":
		fmt.Fprintln(stdout, "Tasks:")
		fmt.Fprintln(stdout, "  1. Buy groceries [done]")
		fmt.Fprintln(stdout, "  2. Walk dog [pending]")
		return 0
	case "add":
		if len(args) < 2 {
			fmt.Fprintln(stderr, "error: task title required")
			// BUG: returns 0 on error, should return 1
			return 0
		}
		fmt.Fprintf(stdout, "Added: %s\n", args[1])
		return 0
	case "help":
		fmt.Fprintln(stdout, "usage: taskcli <command>")
		fmt.Fprintln(stdout, "commands: list, add, help")
		return 0
	default:
		fmt.Fprintf(stderr, "error: unknown command %q\n", args[0])
		// BUG: returns 0 on error, should return 1
		return 0
	}
}

func main() {
	code := runApp(os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(code)
}
