package main

import "testing"

func TestRunWithPort(t *testing.T) {
	cfg, err := run([]string{"--port", "3000"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 3000 {
		t.Fatalf("expected port 3000, got %d", cfg.Port)
	}
}

func TestRunVerbose(t *testing.T) {
	cfg, err := run([]string{"--verbose"})
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Verbose {
		t.Fatal("expected verbose=true")
	}
}

func TestRunDBPath(t *testing.T) {
	cfg, err := run([]string{"--db", "custom.db"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DBPath != "custom.db" {
		t.Fatalf("expected db=custom.db, got %s", cfg.DBPath)
	}
}
