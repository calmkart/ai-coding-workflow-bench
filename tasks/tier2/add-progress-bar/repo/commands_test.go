package main

import (
	"bytes"
	"testing"
)

func TestProgressBarBasic(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgressBar(10, &buf)
	bar.Update(5)

	if buf.Len() == 0 {
		t.Fatal("expected output")
	}
}
