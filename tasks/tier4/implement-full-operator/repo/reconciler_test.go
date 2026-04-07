package operator

import "testing"

func TestBasicReconcile(t *testing.T) {
	store := NewStore()
	store.Put(&Resource{
		Metadata: Metadata{Name: "test", Generation: 1},
		Spec:     ResourceSpec{Replicas: 3, Image: "nginx"},
		Status:   ResourceStatus{Phase: "Pending"},
	})

	rec := NewSimpleReconciler(store)
	result := rec.Reconcile("test")
	if !result.Requeue {
		t.Fatal("should requeue")
	}
}
