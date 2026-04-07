package batcher

import (
	"testing"
	"time"
)

func TestNewBatchProcessor(t *testing.T) {
	bp := NewBatchProcessor[int](10, time.Second, func(batch []int) error {
		return nil
	})
	if bp == nil {
		t.Fatal("expected non-nil")
	}
}
