package statemgr

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Reconciler interface {
	Reconcile(name string) Result
}

type Store struct {
	mu        sync.RWMutex
	resources map[string]*Resource
}

func NewStore() *Store {
	return &Store{resources: make(map[string]*Resource)}
}

func (s *Store) Get(name string) (*Resource, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.resources[name]
	if !ok {
		return nil, false
	}
	cp := *r
	return &cp, true
}

func (s *Store) Put(r *Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r.ResourceVersion++
	s.resources[r.Name] = r
}

// SimpleReconciler implements Reconciler.
// BUG: On error, returns Requeue with no backoff - rapid retry loop.
type SimpleReconciler struct {
	store      *Store
	failUntil  map[string]int // for testing: fail until N-th attempt
}

func NewSimpleReconciler(store *Store) *SimpleReconciler {
	return &SimpleReconciler{
		store:     store,
		failUntil: make(map[string]int),
	}
}

// SetFailUntil configures the reconciler to fail for a resource until the Nth attempt.
func (r *SimpleReconciler) SetFailUntil(name string, attempts int) {
	r.failUntil[name] = attempts
}

func (r *SimpleReconciler) Reconcile(name string) Result {
	res, ok := r.store.Get(name)
	if !ok {
		return Result{Requeue: false}
	}

	// Simulate failures
	if limit, ok := r.failUntil[name]; ok && res.Status.FailCount < limit {
		res.Status.FailCount++
		res.Status.Message = fmt.Sprintf("transient error (attempt %d)", res.Status.FailCount)
		r.store.Put(res)
		// BUG: No backoff - returns fixed short delay regardless of failure count
		return Result{Requeue: true, RequeueAfter: 10 * time.Millisecond, Err: errors.New("transient error")}
	}

	switch res.Status.Phase {
	case "Pending":
		res.Status.Phase = "Running"
		res.Status.FailCount = 0
		r.store.Put(res)
		return Result{Requeue: true, RequeueAfter: 100 * time.Millisecond}
	case "Running":
		if res.Status.ReadyReplicas < res.Spec.Replicas {
			res.Status.ReadyReplicas++
			r.store.Put(res)
			return Result{Requeue: true, RequeueAfter: 100 * time.Millisecond}
		}
		res.Status.Message = "All replicas ready"
		r.store.Put(res)
		return Result{Requeue: false}
	default:
		return Result{Requeue: false}
	}
}

// Runner runs the reconcile loop.
// BUG: Does not apply backoff on errors - uses result.RequeueAfter directly.
type Runner struct {
	reconciler Reconciler
	stopCh     chan struct{}
}

func NewRunner(rec Reconciler) *Runner {
	return &Runner{
		reconciler: rec,
		stopCh:     make(chan struct{}),
	}
}

func (r *Runner) RunOnce(name string, maxIterations int) (int, []time.Duration) {
	var delays []time.Duration
	iterations := 0
	for i := 0; i < maxIterations; i++ {
		result := r.reconciler.Reconcile(name)
		iterations++
		if !result.Requeue {
			break
		}
		// BUG: Uses result.RequeueAfter directly, no exponential backoff on errors
		delays = append(delays, result.RequeueAfter)
		time.Sleep(result.RequeueAfter)
	}
	return iterations, delays
}

func (r *Runner) Stop() {
	close(r.stopCh)
}

func FormatResource(r *Resource) string {
	return fmt.Sprintf("%s: phase=%s replicas=%d/%d fails=%d",
		r.Name, r.Status.Phase, r.Status.ReadyReplicas, r.Spec.Replicas, r.Status.FailCount)
}
