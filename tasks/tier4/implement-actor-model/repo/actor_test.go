package actor

import "testing"

func TestNewActorSystem(t *testing.T) {
	sys := NewActorSystem()
	if sys == nil {
		t.Fatal("system should not be nil")
	}
}
