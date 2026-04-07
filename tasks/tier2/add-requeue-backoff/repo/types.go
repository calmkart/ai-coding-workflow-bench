package statemgr

import "time"

type Resource struct {
	Name            string
	Spec            ResourceSpec
	Status          ResourceStatus
	ResourceVersion int
}

type ResourceSpec struct {
	Replicas int
	Image    string
}

type ResourceStatus struct {
	Phase         string
	ReadyReplicas int
	Message       string
	FailCount     int
}

type Result struct {
	Requeue      bool
	RequeueAfter time.Duration
	Err          error
}
