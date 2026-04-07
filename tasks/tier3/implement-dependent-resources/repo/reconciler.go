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

func (s *Store) List() []*Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Resource, 0, len(s.resources))
	for _, r := range s.resources {
		result = append(result, r)
	}
	return result
}

// AppReconciler only creates the main resource.
// PROBLEM: Does not create dependent resources (Config, Secret).
type AppReconciler struct {
	store *Store
}

func NewAppReconciler(store *Store) *AppReconciler {
	return &AppReconciler{store: store}
}

// Reconcile only handles the main resource.
// PROBLEM: Should create dependent resources in order:
//   1. Secret (no dependencies)
//   2. Config (depends on Secret)
//   3. Main App (depends on Config)
func (r *AppReconciler) Reconcile(name string) Result {
	res, ok := r.store.Get(name)
	if !ok {
		return Result{Requeue: false}
	}

	// PROBLEM: No dependent resource creation
	switch res.Status.Phase {
	case "", "Pending":
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
		res.Status.Message = "All replicas ready"
		r.store.Put(res)
	}

	return Result{Requeue: false}
}
