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

// SimpleReconciler processes resources.
// PROBLEM: No event recording - impossible to track what happened during reconciliation.
type SimpleReconciler struct {
	store *Store
	// MISSING: EventRecorder field
}

func NewSimpleReconciler(store *Store) *SimpleReconciler {
	return &SimpleReconciler{store: store}
}

// Reconcile processes a resource.
// PROBLEM: No events recorded at any step.
func (r *SimpleReconciler) Reconcile(name string) Result {
	res, ok := r.store.Get(name)
	if !ok {
		// PROBLEM: Should record "ResourceNotFound" warning event
		return Result{Requeue: false}
	}

	switch res.Status.Phase {
	case "", "Pending":
		// PROBLEM: Should record "ReconcileStarted" event
		res.Status.Phase = "Running"
		res.Status.Message = "Starting"
		r.store.Put(res)
		// PROBLEM: Should record "StatusChanged" event
		return Result{Requeue: true, RequeueAfter: 50 * time.Millisecond}

	case "Running":
		if res.Status.ReadyReplicas < res.Spec.Replicas {
			res.Status.ReadyReplicas++
			r.store.Put(res)
			// PROBLEM: Should record "ReplicaReady" event
			return Result{Requeue: true, RequeueAfter: 50 * time.Millisecond}
		}
		res.Status.Message = "All replicas ready"
		r.store.Put(res)
		// PROBLEM: Should record "ReconcileComplete" event
	}

	return Result{Requeue: false}
}
