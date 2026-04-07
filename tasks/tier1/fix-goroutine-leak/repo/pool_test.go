package workerpool

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolExecutesWork(t *testing.T) {
	pool := NewPool(2)
	pool.Start()

	var count atomic.Int32
	for i := 0; i < 10; i++ {
		pool.Submit(func() {
			count.Add(1)
		})
	}

	// Wait for work to complete
	time.Sleep(100 * time.Millisecond)

	if count.Load() != 10 {
		t.Fatalf("expected 10 tasks completed, got %d", count.Load())
	}
}
