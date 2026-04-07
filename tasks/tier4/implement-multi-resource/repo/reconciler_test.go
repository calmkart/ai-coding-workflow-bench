package multireconciler

import "testing"

func TestSingleReconciler(t *testing.T) {
	store := NewStore()
	store.PutApplication(&Application{
		Metadata: Metadata{Name: "test-app", Kind: "Application", UID: "uid-1"},
		Spec:     ApplicationSpec{ServicePort: 8080, Replicas: 1, Image: "nginx"},
		Status:   ApplicationStatus{Phase: "Pending"},
	})

	rec := NewSingleReconciler(store)
	rec.Reconcile("test-app")

	app, _ := store.GetApplication("test-app")
	// Single reconciler just sets Ready without dependents
	if app.Status.Phase != "Ready" {
		t.Fatalf("expected Ready, got %s", app.Status.Phase)
	}
}
