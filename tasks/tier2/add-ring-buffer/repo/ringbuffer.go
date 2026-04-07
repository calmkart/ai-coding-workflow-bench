package ringbuffer

// RingBuffer is a fixed-size circular buffer.
// TODO: Implement Write, Read, Len, IsFull, IsEmpty, ToSlice.
type RingBuffer[T any] struct {
	data  []T
	cap   int
	head  int
	tail  int
	count int
}

// NewRingBuffer creates a new RingBuffer with the given capacity.
func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	if capacity < 1 {
		panic("capacity must be >= 1")
	}
	return &RingBuffer[T]{
		data: make([]T, capacity),
		cap:  capacity,
	}
}
