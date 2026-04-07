package operator

import "time"

// Resource represents a managed resource (Kubernetes-style).
type Resource struct {
	Metadata Metadata
	Spec     ResourceSpec
	Status   ResourceStatus
}

type Metadata struct {
	Name       string
	Generation int64
	Labels     map[string]string
}

type ResourceSpec struct {
	Replicas int
	Image    string
}

// ResourceStatus is incomplete — missing Conditions, ObservedGeneration.
// TODO: Add Conditions []Condition, ObservedGeneration int64
type ResourceStatus struct {
	Phase         string // "Pending", "Progressing", "Running", "Degraded", "Failed"
	ReadyReplicas int
	Message       string
}

// Condition represents a resource condition.
// TODO: Define and use in Status.
type Condition struct {
	Type               string    // "Ready", "Available", "Progressing", "Degraded"
	Status             string    // "True", "False", "Unknown"
	LastTransitionTime time.Time
	Reason             string
	Message            string
}

// Result is returned by Reconcile.
// TODO: Add RequeueAfter time.Duration
type Result struct {
	Requeue bool
}
