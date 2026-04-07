package scheduler

import "testing"

func TestNewScheduler(t *testing.T) {
	s := NewScheduler()
	if s == nil {
		t.Fatal("scheduler should not be nil")
	}
}
