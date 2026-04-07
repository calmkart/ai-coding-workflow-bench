package statemgr

import (
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
	return r, ok
}

func (s *Store) Put(r *Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[r.Name] = r
}

type SimpleReconciler struct {
	store *Store
}

func NewSimpleReconciler(store *Store) *SimpleReconciler {
	return &SimpleReconciler{store: store}
}

// Reconcile uses Phase string to track status.
// PROBLEM: Should use Conditions array instead.
func (r *SimpleReconciler) Reconcile(name string) Result {
	res, ok := r.store.Get(name)
	if !ok {
		return Result{Requeue: false}
	}

	switch res.Status.Phase {
	case "", "Pending":
		// PROBLEM: Just setting Phase, should set conditions
		res.Status.Phase = "Running"
		res.Status.Message = "Starting"
		r.store.Put(res)
		return Result{Requeue: true, RequeueAfter: 50 * time.Millisecond}

	case "Running":
		if res.Status.ReadyReplicas < res.Spec.Replicas {
			res.Status.ReadyReplicas++
			r.store.Put(res)
			return Result{Requeue: true, RequeueAfter: 50 * time.Millisecond}
		}
		// PROBLEM: Should set Ready condition to True
		res.Status.Message = "All replicas ready"
		r.store.Put(res)
	}

	return Result{Requeue: false}
}
