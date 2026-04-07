package statemgr

import "testing"

func TestBasic(t *testing.T) {
	store := NewStore()
	store.Put(&Resource{
		Name: "app1",
		Kind: "App",
		Spec: ResourceSpec{Replicas: 1, Image: "nginx"},
		Status: ResourceStatus{Phase: "Pending"},
	})
	rec := NewAppReconciler(store)
	rec.Reconcile("app1")
}
