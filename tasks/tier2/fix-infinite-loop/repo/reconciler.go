package statemgr

import (
	"fmt"
	"sync"
	"time"
)

// Reconciler processes a single resource.
type Reconciler interface {
	Reconcile(name string) Result
}

// Store holds resources.
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
	return r, ok
}

func (s *Store) Put(r *Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[name(r)] = r
}

func name(r *Resource) string {
	return r.Name
}

// SimpleReconciler implements Reconciler.
type SimpleReconciler struct {
	store *Store
}

func NewSimpleReconciler(store *Store) *SimpleReconciler {
	return &SimpleReconciler{store: store}
}

func (r *SimpleReconciler) Reconcile(name string) Result {
	res, ok := r.store.Get(name)
	if !ok {
		return Result{Requeue: false}
	}

	switch res.Status.Phase {
	case "Pending":
		res.Status.Phase = "Running"
		res.Status.Message = "Starting"
		r.store.Put(res)
		return Result{Requeue: true} // BUG: No delay
	case "Running":
		if res.Status.ReadyReplicas < res.Spec.Replicas {
			res.Status.ReadyReplicas++
			r.store.Put(res)
			return Result{Requeue: true} // BUG: No delay - causes tight loop
		}
		res.Status.Message = "All replicas ready"
		r.store.Put(res)
		return Result{Requeue: false}
	default:
		return Result{Requeue: false}
	}
}

// Runner runs the reconcile loop.
// BUG: When Requeue is true, immediately reconciles again with no delay.
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

// RunOnce reconciles a resource and returns the number of iterations.
func (r *Runner) RunOnce(name string, maxIterations int) int {
	iterations := 0
	for i := 0; i < maxIterations; i++ {
		result := r.reconciler.Reconcile(name)
		iterations++
		if !result.Requeue {
			break
		}
		// BUG: No sleep/delay before retry - tight loop
	}
	return iterations
}

// Run continuously reconciles until stopped.
func (r *Runner) Run(name string) {
	for {
		select {
		case <-r.stopCh:
			return
		default:
			result := r.reconciler.Reconcile(name)
			if !result.Requeue {
				return
			}
			// BUG: No delay - CPU spin
		}
	}
}

func (r *Runner) Stop() {
	close(r.stopCh)
}

// Timestamp helper
func now() string {
	return time.Now().Format(time.RFC3339)
}

// Format resource info
func FormatResource(r *Resource) string {
	return fmt.Sprintf("%s: phase=%s replicas=%d/%d msg=%s",
		r.Name, r.Status.Phase, r.Status.ReadyReplicas, r.Spec.Replicas, r.Status.Message)
}
