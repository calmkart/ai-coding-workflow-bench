package processor

import "testing"

func TestProcessAllNormal(t *testing.T) {
	results, err := ProcessAll([]string{"hello"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0] != "HELLO" {
		t.Fatalf("expected [HELLO], got %v", results)
	}
}
