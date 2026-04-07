package workerpool

// Pool is a fixed-size worker pool.
// TODO: Implement Submit, Wait, Shutdown.
type Pool struct {
	workers int
	tasks   chan func()
	done    chan struct{}
}

// NewPool creates a new worker pool with the given number of workers.
func NewPool(workers int) *Pool {
	if workers < 1 {
		panic("workers must be >= 1")
	}
	p := &Pool{
		workers: workers,
		tasks:   make(chan func(), 100),
		done:    make(chan struct{}),
	}
	// TODO: start worker goroutines
	return p
}
