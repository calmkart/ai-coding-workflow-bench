package multireconciler

// SingleReconciler only handles Application — no cascading to Service or Endpoint.
// BUG: Does not create dependent resources.
// BUG: Does not propagate status from dependents.
type SingleReconciler struct {
	store *Store
}

func NewSingleReconciler(store *Store) *SingleReconciler {
	return &SingleReconciler{store: store}
}

func (r *SingleReconciler) Reconcile(name string) Result {
	app, ok := r.store.GetApplication(name)
	if !ok {
		return Result{Requeue: false}
	}

	// BUG: Just sets status to Ready without actually creating dependents.
	// Should: create Service → create Endpoint → check Endpoint Ready → set Ready
	switch app.Status.Phase {
	case "Pending":
		app.Status.Phase = "Ready" // BUG: skips creating Service/Endpoint
		app.Status.Message = "Ready (but no dependents created!)"
		r.store.PutApplication(app)
		return Result{Requeue: false}
	default:
		return Result{Requeue: false}
	}
}

// TODO: Implement MultiReconciler that:
// 1. Creates Service when Application is created
// 2. Creates Endpoint when Service is created
// 3. Propagates Ready status upward
// 4. Cascades deletion downward
// 5. Updates dependents when spec changes

// TODO: Implement ResourceGraph that defines:
// Application → Service → Endpoint dependency chain
