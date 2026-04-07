package circuitbreaker

import (
	"errors"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed   State = iota // Normal operation
	StateOpen                  // Circuit is open, rejecting calls
	StateHalfOpen              // Allowing limited trial calls
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// ErrCircuitOpen is returned when the circuit breaker is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// Options configures the circuit breaker.
type Options struct {
	MaxFailures     int           // Consecutive failures before opening
	Timeout         time.Duration // How long to stay open before half-open
	HalfOpenMaxCalls int          // Max calls allowed in half-open state
}

// CircuitBreaker implements the circuit breaker pattern.
// TODO: Implement the state machine:
//   - Closed: execute normally, count failures
//   - Open: reject all calls with ErrCircuitOpen
//   - HalfOpen: allow limited trial calls
type CircuitBreaker struct {
	opts     Options
	state    State
	failures int
	lastFail time.Time
}

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(opts Options) *CircuitBreaker {
	if opts.MaxFailures <= 0 {
		opts.MaxFailures = 5
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second
	}
	if opts.HalfOpenMaxCalls <= 0 {
		opts.HalfOpenMaxCalls = 1
	}
	return &CircuitBreaker{
		opts:  opts,
		state: StateClosed,
	}
}

// Execute runs the given function through the circuit breaker.
// TODO: Implement state machine logic.
// Currently just calls fn directly with no protection.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	// PROBLEM: No circuit breaker logic, just passthrough
	return fn()
}

// State returns the current state.
func (cb *CircuitBreaker) State() State {
	return cb.state
}
