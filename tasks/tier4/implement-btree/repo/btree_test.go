package btree

import "testing"

func TestNewBTree(t *testing.T) {
	tree := NewBTree[int, string](3, func(a, b int) bool { return a < b })
	if tree.Len() != 0 {
		t.Fatalf("new tree should be empty, got %d", tree.Len())
	}
}
