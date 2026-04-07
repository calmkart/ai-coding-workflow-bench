package statemgr

import "time"

type Resource struct {
	Name   string
	Spec   ResourceSpec
	Status ResourceStatus
}

type ResourceSpec struct {
	Replicas int
	Image    string
}

// ResourceStatus uses a single Phase string.
// PROBLEM: Cannot express multiple parallel conditions.
// PROBLEM: No timestamps for state transitions.
// Need to add Conditions []Condition array.
type ResourceStatus struct {
	Phase         string // PROBLEM: Only one state at a time
	ReadyReplicas int
	Message       string
	// MISSING: Conditions []Condition
}

type Result struct {
	Requeue      bool
	RequeueAfter time.Duration
}
