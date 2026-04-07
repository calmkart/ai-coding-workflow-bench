// Package workerpool provides concurrency utilities.
package workerpool

// Counter is a simple counter.
// BUG: not safe for concurrent use - no synchronization on value field.
type Counter struct {
	value int64
}

// NewCounter creates a new Counter starting at 0.
func NewCounter() *Counter {
	return &Counter{}
}

// Inc increments the counter by 1.
// BUG: not safe for concurrent use - direct field access without lock.
func (c *Counter) Inc() {
	c.value++
}

// Get returns the current counter value.
// BUG: not safe for concurrent use - direct field access without lock.
func (c *Counter) Get() int64 {
	return c.value
}

// Reset resets the counter to 0.
func (c *Counter) Reset() {
	c.value = 0
}
