package eventbus

import "testing"

func TestNewEventBus(t *testing.T) {
	eb := NewEventBus()
	if eb == nil {
		t.Fatal("expected non-nil event bus")
	}
}
