package processor

import "testing"

func TestProcessSuccess(t *testing.T) {
	p := NewProcessor("test")
	result := p.Process(Item{ID: 1, Name: "item1", Data: "hello"})

	if !result.Success {
		t.Fatal("expected success")
	}
	if result.Output != "test:hello" {
		t.Fatalf("expected 'test:hello', got '%s'", result.Output)
	}
}

func TestProcessEmptyData(t *testing.T) {
	p := NewProcessor("test")
	result := p.Process(Item{ID: 1, Name: "item1", Data: ""})

	if result.Success {
		t.Fatal("expected failure for empty data")
	}
}
