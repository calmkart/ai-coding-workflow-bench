package pubsub

import "testing"

func TestNewPubSub(t *testing.T) {
	ps := NewPubSub()
	if ps == nil {
		t.Fatal("expected non-nil")
	}
}
