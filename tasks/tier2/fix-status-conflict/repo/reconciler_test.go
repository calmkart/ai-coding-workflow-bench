package statemgr

import "testing"

func TestNewStore(t *testing.T) {
	s := NewStore()
	if s == nil {
		t.Fatal("expected non-nil")
	}
}

func TestPutGet(t *testing.T) {
	s := NewStore()
	r := &Resource{Name: "test", Spec: ResourceSpec{Replicas: 1}}
	s.Put(r)

	got, ok := s.Get("test")
	if !ok {
		t.Fatal("expected resource")
	}
	if got.Name != "test" {
		t.Fatalf("expected 'test', got '%s'", got.Name)
	}
}
