package statemgr

import (
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
	return r, ok
}

func (s *Store) Put(r *Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[r.Name] = r
}

// Delete removes a resource.
// PROBLEM: No cascade deletion of child resources.
func (s *Store) Delete(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.resources[name]
	if ok {
		delete(s.resources, name)
	}
	return ok
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

type AppReconciler struct {
	store *Store
}

func NewAppReconciler(store *Store) *AppReconciler {
	return &AppReconciler{store: store}
}

// Reconcile processes a resource.
// PROBLEM: Creates child resources but does not set owner reference.
func (r *AppReconciler) Reconcile(name string) Result {
	res, ok := r.store.Get(name)
	if !ok {
		return Result{Requeue: false}
	}

	switch res.Status.Phase {
	case "", "Pending":
		// Create child resources
		// PROBLEM: No owner reference set on children
		child := &Resource{
			Name:   fmt.Sprintf("%s-child-1", name),
			Kind:   "Child",
			Spec:   ResourceSpec{Replicas: 1, Image: res.Spec.Image},
			Status: ResourceStatus{Phase: "Pending"},
		}
		r.store.Put(child)

		res.Status.Phase = "Running"
		res.Status.Message = "Children created"
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
