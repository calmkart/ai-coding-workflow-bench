package semaphore

// Semaphore is a counting semaphore based on a buffered channel.
// TODO: Implement Acquire, TryAcquire, Release.
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore creates a semaphore with n permits.
func NewSemaphore(n int) *Semaphore {
	if n < 1 {
		panic("semaphore count must be >= 1")
	}
	return &Semaphore{
		ch: make(chan struct{}, n),
	}
}
