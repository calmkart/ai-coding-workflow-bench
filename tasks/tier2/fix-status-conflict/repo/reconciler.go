package statemgr

import (
	"fmt"
	"sync"
	"time"
)

type Reconciler interface {
	Reconcile(name string) Result
}

// Store holds resources.
// BUG: Put does not check version - concurrent updates silently overwrite.
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
	// Return a copy
	cp := *r
	return &cp, true
}

// Put stores a resource.
// BUG: No version check - should return ErrConflict if version mismatch.
func (s *Store) Put(r *Resource) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[r.Name] = r
	return nil
}

type SimpleReconciler struct {
	store *Store
}

func NewSimpleReconciler(store *Store) *SimpleReconciler {
	return &SimpleReconciler{store: store}
}

// BUG: Does not handle ErrConflict from Put - should retry on conflict.
func (r *SimpleReconciler) Reconcile(name string) Result {
	res, ok := r.store.Get(name)
	if !ok {
		return Result{Requeue: false}
	}

	switch res.Status.Phase {
	case "Pending":
		res.Status.Phase = "Running"
		res.Status.Message = "Starting"
		r.store.Put(res) // BUG: ignores error
		return Result{Requeue: true, RequeueAfter: 100 * time.Millisecond}
	case "Running":
		if res.Status.ReadyReplicas < res.Spec.Replicas {
			res.Status.ReadyReplicas++
			r.store.Put(res) // BUG: ignores error
			return Result{Requeue: true, RequeueAfter: 100 * time.Millisecond}
		}
		res.Status.Message = "All replicas ready"
		r.store.Put(res) // BUG: ignores error
		return Result{Requeue: false}
	default:
		return Result{Requeue: false}
	}
}

func FormatResource(r *Resource) string {
	return fmt.Sprintf("%s: phase=%s replicas=%d/%d",
		r.Name, r.Status.Phase, r.Status.ReadyReplicas, r.Spec.Replicas)
}
