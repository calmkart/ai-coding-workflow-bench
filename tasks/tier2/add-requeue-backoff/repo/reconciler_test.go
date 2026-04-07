package statemgr

import "testing"

func TestNewStore(t *testing.T) {
	s := NewStore()
	if s == nil {
		t.Fatal("expected non-nil")
	}
}

func TestBasicReconcile(t *testing.T) {
	store := NewStore()
	store.Put(&Resource{
		Name: "test",
		Spec: ResourceSpec{Replicas: 1},
		Status: ResourceStatus{Phase: "Pending"},
	})

	rec := NewSimpleReconciler(store)
	result := rec.Reconcile("test")
	if !result.Requeue {
		t.Fatal("expected requeue")
	}
}
