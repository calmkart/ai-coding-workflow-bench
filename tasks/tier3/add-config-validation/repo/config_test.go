package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	os.WriteFile(path, []byte(`{"data_dir":"/tmp","max_tasks":100}`), 0644)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DataDir != "/tmp" {
		t.Fatalf("expected /tmp, got %s", cfg.DataDir)
	}
}
