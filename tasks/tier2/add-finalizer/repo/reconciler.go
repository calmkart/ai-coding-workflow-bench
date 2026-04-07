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
	// CleanupLog records cleanup actions for testing
	CleanupLog []string
}

func NewStore() *Store {
	return &Store{
		resources:  make(map[string]*Resource),
		CleanupLog: []string{},
	}
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

func (s *Store) Delete(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.resources, name)
}

func (s *Store) LogCleanup(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CleanupLog = append(s.CleanupLog, msg)
}

// SimpleReconciler implements Reconciler.
// BUG: No finalizer handling - deletion skips cleanup.
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

	// BUG: No finalizer check - should add finalizer on creation
	// BUG: No deletion handling - should run cleanup before deletion

	switch res.Status.Phase {
	case "Pending":
		res.Status.Phase = "Running"
		res.Status.Message = "Starting"
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

func FormatResource(r *Resource) string {
	return fmt.Sprintf("%s: phase=%s replicas=%d/%d",
		r.Name, r.Status.Phase, r.Status.ReadyReplicas, r.Spec.Replicas)
}
