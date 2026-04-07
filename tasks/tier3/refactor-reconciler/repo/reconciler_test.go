package statemgr

import "testing"

func TestGodReconciler(t *testing.T) {
	store := NewStore()
	store.Put(&Resource{
		Name: "test",
		Spec: ResourceSpec{Replicas: 2, Image: "nginx"},
		Status: ResourceStatus{Phase: "Pending"},
	})

	rec := NewGodReconciler(store)
	result := rec.Reconcile("test")
	if !result.Requeue {
		t.Fatal("expected requeue")
	}
}
