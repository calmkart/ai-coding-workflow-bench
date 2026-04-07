package main

import (
	"bytes"
	"testing"
)

func TestListCmd(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runApp([]string{"list"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if stdout.Len() == 0 {
		t.Fatal("expected output from list command")
	}
}

func TestHelpCmd(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runApp([]string{"help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestAddCmd(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runApp([]string{"add", "test task"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}
