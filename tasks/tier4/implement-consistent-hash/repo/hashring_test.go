package hashring

import "testing"

func TestNewHashRing(t *testing.T) {
	ring := NewHashRing(100)
	if ring.NodeCount() != 0 {
		t.Fatalf("new ring should have 0 nodes, got %d", ring.NodeCount())
	}
}
