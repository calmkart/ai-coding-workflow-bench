package coordinator

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Task represents a unit of work.
type Task struct {
	ID      int
	Payload string
}

// Result holds the result of processing a task.
type Result struct {
	TaskID int
	Output string
	Err    error
}

// Worker processes tasks.
type Worker struct {
	id       int
	process  func(Task) Result
	tasks    <-chan Task
	results  chan<- Result
	running  atomic.Bool
}

// Coordinator manages multiple workers.
// BUG: Context cancellation does not propagate to workers.
type Coordinator struct {
	workers    []*Worker
	workerCount int
	process    func(Task) Result
	tasks      chan Task
	results    chan Result
	wg         sync.WaitGroup
}

// NewCoordinator creates a new Coordinator.
func NewCoordinator(workerCount int, process func(Task) Result) *Coordinator {
	return &Coordinator{
		workerCount: workerCount,
		process:     process,
		tasks:       make(chan Task, 100),
		results:     make(chan Result, 100),
	}
}

// Start launches all workers.
// BUG: Workers use context.Background() instead of the passed ctx.
// BUG: Workers don't check ctx.Done() in their processing loop.
func (c *Coordinator) Start(ctx context.Context) {
	for i := 0; i < c.workerCount; i++ {
		w := &Worker{
			id:      i,
			process: c.process,
			tasks:   c.tasks,
			results: c.results,
		}
		c.workers = append(c.workers, w)
		c.wg.Add(1)

		// BUG: Uses context.Background() instead of ctx
		go func(w *Worker) {
			defer c.wg.Done()
			w.running.Store(true)
			defer w.running.Store(false)

			// BUG: context.Background() - should use ctx
			bgCtx := context.Background()
			_ = bgCtx

			for task := range w.tasks {
				// BUG: No check for ctx.Done()
				// This worker will continue processing even after ctx is cancelled
				result := w.process(task)
				w.results <- result
			}
		}(w)
	}
}

// Submit adds a task.
// BUG: No context check - can submit after cancellation.
func (c *Coordinator) Submit(task Task) {
	c.tasks <- task
}

// Results returns the results channel.
func (c *Coordinator) Results() <-chan Result {
	return c.results
}

// Wait waits for all workers to finish.
// BUG: Only closes tasks channel, doesn't propagate cancellation.
func (c *Coordinator) Wait() {
	close(c.tasks)
	c.wg.Wait()
	close(c.results)
}

// WorkerCount returns the number of running workers.
func (c *Coordinator) WorkerCount() int {
	count := 0
	for _, w := range c.workers {
		if w.running.Load() {
			count++
		}
	}
	return count
}

// SetWorkerTimeout is a no-op placeholder.
// TODO: Add per-worker timeout support.
func (c *Coordinator) SetWorkerTimeout(d time.Duration) {
	// PROBLEM: Not implemented
}
