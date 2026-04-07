package statemgr

// Resource represents a managed resource.
type Resource struct {
	Name   string
	Spec   ResourceSpec
	Status ResourceStatus
}

type ResourceSpec struct {
	Replicas int
	Image    string
}

type ResourceStatus struct {
	Phase          string // "Pending", "Running", "Failed"
	ReadyReplicas  int
	Message        string
}

// Result is returned by Reconcile to indicate what should happen next.
// BUG: No RequeueAfter field - Requeue:true causes immediate retry.
type Result struct {
	Requeue bool
}
