package multireconciler

import "time"

// Result is returned by Reconcile.
type Result struct {
	Requeue      bool
	RequeueAfter time.Duration
}

// OwnerReference points to the owner of this resource.
type OwnerReference struct {
	Kind string
	Name string
	UID  string
}

// Metadata holds resource metadata.
type Metadata struct {
	Name            string
	Kind            string
	UID             string
	OwnerReferences []OwnerReference
}

// Application is the top-level resource.
type Application struct {
	Metadata Metadata
	Spec     ApplicationSpec
	Status   ApplicationStatus
}

type ApplicationSpec struct {
	ServicePort int
	Replicas    int
	Image       string
}

type ApplicationStatus struct {
	Phase   string // Pending, Progressing, Ready, Failed
	Message string
}

// Service is created by Application.
type Service struct {
	Metadata Metadata
	Spec     ServiceSpec
	Status   ServiceStatus
}

type ServiceSpec struct {
	Port       int
	TargetPort int
	AppName    string
}

type ServiceStatus struct {
	Phase   string
	Message string
}

// Endpoint is created by Service.
type Endpoint struct {
	Metadata Metadata
	Spec     EndpointSpec
	Status   EndpointStatus
}

type EndpointSpec struct {
	Address     string
	Port        int
	ServiceName string
}

type EndpointStatus struct {
	Phase   string // Pending, Ready
	Message string
}
