package main

import (
	"flag"
	"fmt"
	"os"
)

// Config holds the CLI configuration.
type Config struct {
	Port    int
	Verbose bool
	DBPath  string
}

// run parses arguments and returns a Config.
// This function signature is part of the API contract and must not be changed.
func run(args []string) (*Config, error) {
	fs := flag.NewFlagSet("taskcli", flag.ContinueOnError)

	// BUG: default port is 0, should be 8080
	port := fs.Int("port", 0, "server port")
	verbose := fs.Bool("verbose", false, "enable verbose logging")
	dbPath := fs.String("db", "tasks.db", "database file path")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return &Config{
		Port:    *port,
		Verbose: *verbose,
		DBPath:  *dbPath,
	}, nil
}

func main() {
	cfg, err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Starting taskcli on port %d\n", cfg.Port)
}
