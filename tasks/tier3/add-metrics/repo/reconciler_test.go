package statemgr

import "testing"

func TestBasic(t *testing.T) {
	store := NewStore()
	store.Put(&Resource{
		Name:   "test",
		Spec:   ResourceSpec{Replicas: 1, Image: "nginx"},
		Status: ResourceStatus{Phase: "Pending"},
	})
	rec := NewSimpleReconciler(store)
	rec.Reconcile("test")
}
