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
	fmt.Println("taskcli")
}
