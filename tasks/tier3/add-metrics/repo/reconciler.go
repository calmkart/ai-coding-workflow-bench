package statemgr

import (
	"sync"
	"time"
)

// Reconciler processes a single resource.
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
	return r, ok
}

func (s *Store) Put(r *Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[r.Name] = r
}

// SimpleReconciler has no metrics collection.
// PROBLEM: Cannot observe reconcile count, latency, or error rate.
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
	case "", "Pending":
		res.Status.Phase = "Running"
		res.Status.Message = "Starting"
		r.store.Put(res)
		return Result{Requeue: true, RequeueAfter: 10 * time.Millisecond}

	case "Running":
		if res.Status.ReadyReplicas < res.Spec.Replicas {
			res.Status.ReadyReplicas++
			r.store.Put(res)
			return Result{Requeue: true, RequeueAfter: 10 * time.Millisecond}
		}
		res.Status.Message = "All replicas ready"
		r.store.Put(res)
	}

	return Result{Requeue: false}
}

// Runner runs the reconcile loop.
type Runner struct {
	reconciler Reconciler
}

func NewRunner(rec Reconciler) *Runner {
	return &Runner{reconciler: rec}
}

func (r *Runner) RunOnce(name string, maxIterations int) int {
	iterations := 0
	for i := 0; i < maxIterations; i++ {
		result := r.reconciler.Reconcile(name)
		iterations++
		if !result.Requeue {
			break
		}
		if result.RequeueAfter > 0 {
			time.Sleep(result.RequeueAfter)
		}
	}
	return iterations
}
