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
	s.resources[r.Name] = r
}

// GodReconciler does everything in one giant function.
// SMELL: This should be split into sub-reconcilers:
//   1. ValidateSpec - validate the resource spec
//   2. EnsureResources - ensure child resources exist
//   3. UpdateStatus - update the resource status
type GodReconciler struct {
	store *Store
}

func NewGodReconciler(store *Store) *GodReconciler {
	return &GodReconciler{store: store}
}

// Reconcile does everything: validate, ensure, update status.
// SMELL: ~100 lines doing three different things.
func (r *GodReconciler) Reconcile(name string) Result {
	res, ok := r.store.Get(name)
	if !ok {
		return Result{Requeue: false}
	}

	// --- PHASE 1: Validate Spec ---
	// SMELL: Should be separate ValidateSpec sub-reconciler
	if res.Spec.Replicas < 0 {
		res.Status.Phase = "Failed"
		res.Status.Message = "invalid replicas: must be >= 0"
		r.store.Put(res)
		return Result{Requeue: false}
	}
	if res.Spec.Image == "" {
		res.Status.Phase = "Failed"
		res.Status.Message = "invalid spec: image is required"
		r.store.Put(res)
		return Result{Requeue: false}
	}
	if res.Spec.Replicas > 10 {
		res.Status.Phase = "Failed"
		res.Status.Message = "invalid replicas: max 10"
		r.store.Put(res)
		return Result{Requeue: false}
	}

	// --- PHASE 2: Ensure Resources ---
	// SMELL: Should be separate EnsureResources sub-reconciler
	switch res.Status.Phase {
	case "", "Pending":
		res.Status.Phase = "Pending"
		res.Status.Message = "Starting deployment"
		res.Status.ReadyReplicas = 0
		r.store.Put(res)
		// Simulate resource creation
		res.Status.Phase = "Running"
		res.Status.Message = "Deploying replicas"
		r.store.Put(res)
		return Result{Requeue: true, RequeueAfter: 100 * time.Millisecond}

	case "Running":
		if res.Status.ReadyReplicas < res.Spec.Replicas {
			res.Status.ReadyReplicas++
			res.Status.Message = fmt.Sprintf("Scaling: %d/%d ready",
				res.Status.ReadyReplicas, res.Spec.Replicas)
			r.store.Put(res)
			return Result{Requeue: true, RequeueAfter: 50 * time.Millisecond}
		}
	}

	// --- PHASE 3: Update Status ---
	// SMELL: Should be separate UpdateStatus sub-reconciler
	if res.Status.ReadyReplicas == res.Spec.Replicas {
		res.Status.Phase = "Running"
		res.Status.Message = "All replicas ready"
		r.store.Put(res)
		return Result{Requeue: false}
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
