package pool

// Pool is a generic object pool.
// TODO: Implement bounded pool with Get/Put, reset, and stats.
// Currently just a skeleton that always creates new objects.

// PoolStats tracks pool usage.
type PoolStats struct {
	Gets     int64 // Total Get calls
	Puts     int64 // Total Put calls
	News     int64 // Objects created via factory
	Discards int64 // Objects discarded (pool full)
}

// PoolOption configures the pool.
type PoolOption func(*poolConfig)

type poolConfig struct {
	maxSize   int
	resetFunc func(any)
}

// Pool is a generic object pool.
// PROBLEM: No actual pooling, just creates new objects every time.
type Pool[T any] struct {
	factory func() T
	config  poolConfig
}

// NewPool creates a new object pool.
func NewPool[T any](factory func() T, opts ...PoolOption) *Pool[T] {
	cfg := poolConfig{maxSize: 10}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Pool[T]{
		factory: factory,
		config:  cfg,
	}
}

// Get retrieves an object from the pool.
// PROBLEM: Always creates a new object, never reuses.
func (p *Pool[T]) Get() T {
	return p.factory()
}

// Put returns an object to the pool.
// PROBLEM: Does nothing, object is not stored for reuse.
func (p *Pool[T]) Put(obj T) {
	// PROBLEM: no-op
}

// Stats returns pool usage statistics.
// PROBLEM: Always returns zeros.
func (p *Pool[T]) Stats() PoolStats {
	return PoolStats{}
}
