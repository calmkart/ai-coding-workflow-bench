package gc

import "time"

// OwnerReference identifies the owner of a resource.
type OwnerReference struct {
	Name string
	Kind string
	UID  string
}

// Metadata holds resource metadata.
type Metadata struct {
	Name              string
	UID               string
	Kind              string
	OwnerReferences   []OwnerReference
	Finalizers        []string
	DeletionTimestamp *time.Time // nil = not deleted
}

// Resource represents a managed resource with owner tracking.
type Resource struct {
	Metadata Metadata
	Spec     map[string]string
}

// IsDeleting returns true if the resource has been marked for deletion.
func (r *Resource) IsDeleting() bool {
	return r.Metadata.DeletionTimestamp != nil
}

// HasFinalizers returns true if there are pending finalizers.
func (r *Resource) HasFinalizers() bool {
	return len(r.Metadata.Finalizers) > 0
}
