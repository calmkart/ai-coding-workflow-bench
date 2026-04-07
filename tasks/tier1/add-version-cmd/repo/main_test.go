package main

import (
	"bytes"
	"testing"
)

func TestListCmd(t *testing.T) {
	var buf bytes.Buffer
	err := runCmd([]string{"list"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Fatal("expected output from list command")
	}
}

func TestAddCmd(t *testing.T) {
	var buf bytes.Buffer
	err := runCmd([]string{"add", "buy milk"}, &buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUnknownCmd(t *testing.T) {
	var buf bytes.Buffer
	err := runCmd([]string{"foobar"}, &buf)
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}
