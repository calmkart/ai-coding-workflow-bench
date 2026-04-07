package batcher

import "time"

// BatchProcessor collects items and processes them in batches.
// TODO: Implement batch collection and flush logic.
type BatchProcessor[T any] struct {
	batchSize     int
	flushInterval time.Duration
	process       func([]T) error
}

// NewBatchProcessor creates a new batch processor.
// PROBLEM: No actual batching logic - items are not collected.
func NewBatchProcessor[T any](batchSize int, flushInterval time.Duration, process func([]T) error) *BatchProcessor[T] {
	return &BatchProcessor[T]{
		batchSize:     batchSize,
		flushInterval: flushInterval,
		process:       process,
	}
}

// Add adds an item to the current batch.
// PROBLEM: Not implemented.
func (bp *BatchProcessor[T]) Add(item T) {
	// PROBLEM: Item is not stored anywhere
}

// Flush processes the current batch immediately.
// PROBLEM: Not implemented.
func (bp *BatchProcessor[T]) Flush() error {
	return nil
}

// Close flushes remaining items and shuts down.
// PROBLEM: Not implemented.
func (bp *BatchProcessor[T]) Close() error {
	return nil
}
