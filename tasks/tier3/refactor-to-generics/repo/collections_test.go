package collections

import "testing"

func TestMapBasic(t *testing.T) {
	input := []interface{}{1, 2, 3}
	result := Map(input, func(v interface{}) interface{} {
		return v.(int) * 2
	})
	if len(result) != 3 {
		t.Fatalf("expected 3, got %d", len(result))
	}
	if result[0].(int) != 2 {
		t.Fatalf("expected 2, got %v", result[0])
	}
}
