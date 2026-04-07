package statemgr

import "time"

type Resource struct {
	Name   string
	Kind   string
	Spec   ResourceSpec
	Status ResourceStatus
	// PROBLEM: No OwnerReference - child resources are not linked to parents.
	// Need to add OwnerRef field and cascade deletion logic.
}

type ResourceSpec struct {
	Replicas int
	Image    string
}

type ResourceStatus struct {
	Phase         string
	ReadyReplicas int
	Message       string
}

type Result struct {
	Requeue      bool
	RequeueAfter time.Duration
}
