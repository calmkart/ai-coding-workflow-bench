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

type ResourceStatus struct {
	Phase         string
	ReadyReplicas int
	Message       string
}

type Result struct {
	Requeue      bool
	RequeueAfter time.Duration
}
