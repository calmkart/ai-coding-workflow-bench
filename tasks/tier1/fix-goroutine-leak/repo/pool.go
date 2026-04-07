// Package workerpool provides a simple worker pool implementation.
package workerpool

// Pool is a worker pool that processes submitted functions concurrently.
type Pool struct {
	size int
	jobs chan func()
}

// NewPool creates a new worker pool with the given number of workers.
func NewPool(size int) *Pool {
	return &Pool{
		size: size,
		jobs: make(chan func(), 100),
	}
}

// Start launches the worker goroutines.
func (p *Pool) Start() {
	for i := 0; i < p.size; i++ {
		go p.worker()
	}
}

// worker processes jobs from the channel.
// BUG: no way to stop this goroutine - it blocks forever on p.jobs channel
// even after Stop is called, because there's no quit signal.
func (p *Pool) worker() {
	for fn := range p.jobs {
		fn()
	}
}

// Stop signals all workers to stop.
// BUG: doesn't actually stop the workers - they keep blocking on the jobs channel.
// The jobs channel is never closed, so workers leak.
func (p *Pool) Stop() {
	// BUG: this is a no-op. Workers are never notified to stop.
	// Should close the jobs channel or use a done channel/context.
}

// Submit sends a function to be executed by a worker.
func (p *Pool) Submit(fn func()) {
	p.jobs <- fn
}
