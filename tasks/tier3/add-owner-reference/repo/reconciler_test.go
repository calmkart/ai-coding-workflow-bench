package statemgr

import "testing"

func TestBasicReconcile(t *testing.T) {
	store := NewStore()
	store.Put(&Resource{
		Name: "app1",
		Kind: "App",
		Spec: ResourceSpec{Replicas: 1, Image: "nginx"},
		Status: ResourceStatus{Phase: "Pending"},
	})

	rec := NewAppReconciler(store)
	result := rec.Reconcile("app1")
	if !result.Requeue {
		t.Fatal("expected requeue")
	}
}
