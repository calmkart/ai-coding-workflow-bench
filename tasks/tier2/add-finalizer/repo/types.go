package statemgr

import "time"

// Resource represents a managed resource.
type Resource struct {
	Name              string
	Spec              ResourceSpec
	Status            ResourceStatus
	// BUG: No Finalizers field - resources deleted without cleanup
	// BUG: No DeletionTimestamp - can't track deletion in progress
}

type ResourceSpec struct {
	Replicas int
	Image    string
}

type ResourceStatus struct {
	Phase         string // "Pending", "Running", "Failed", "Terminating"
	ReadyReplicas int
	Message       string
}

type Result struct {
	Requeue      bool
	RequeueAfter time.Duration
}
