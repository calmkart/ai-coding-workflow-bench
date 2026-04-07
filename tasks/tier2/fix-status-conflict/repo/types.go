package statemgr

import (
	"errors"
	"time"
)

var ErrConflict = errors.New("conflict: resource version mismatch")

// Resource represents a managed resource.
// BUG: No ResourceVersion - updates can silently overwrite concurrent changes.
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
