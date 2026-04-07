package leaderelection

import "testing"

func TestNewLeaderElector(t *testing.T) {
	store := NewInMemoryLeaseStore()
	le := NewLeaderElector("node-1", store, LeaderOpts{})
	if le.IsLeader() {
		t.Fatal("should not be leader initially")
	}
}
