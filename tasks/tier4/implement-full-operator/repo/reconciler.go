package operator

import (
	"sync"
)

// Reconciler processes a single resource.
type Reconciler interface {
	Reconcile(name string) Result
}

// Store holds resources.
// TODO: Add List, Watch, proper CRUD.
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
	s.resources[r.Metadata.Name] = r
}

// SimpleReconciler is incomplete — no Conditions, no events, limited state machine.
// TODO: Implement full state machine with Conditions and events.
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

	// Very basic — no Conditions, no events, no proper state machine
	switch res.Status.Phase {
	case "Pending":
		res.Status.Phase = "Running"
		res.Status.Message = "Started"
		r.store.Put(res)
		return Result{Requeue: true}
	case "Running":
		if res.Status.ReadyReplicas < res.Spec.Replicas {
			res.Status.ReadyReplicas++
			r.store.Put(res)
			return Result{Requeue: true}
		}
		return Result{Requeue: false}
	default:
		return Result{Requeue: false}
	}
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
	}
	return iterations
}
