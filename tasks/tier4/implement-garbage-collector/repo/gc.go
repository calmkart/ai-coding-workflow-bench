package gc

import "context"

// GarbageCollector scans for and removes orphaned resources.
// TODO: Implement CollectOnce and Run.
type GarbageCollector struct {
	store *Store
}

func NewGarbageCollector(store *Store) *GarbageCollector {
	return &GarbageCollector{store: store}
}

// CollectOnce performs one GC cycle.
// Returns the number of resources cleaned up.
// TODO: Implement:
// 1. Find resources whose owners don't exist → mark for deletion
// 2. Find resources marked for deletion with no dependents and no finalizers → purge
func (gc *GarbageCollector) CollectOnce() int {
	return 0
}

// Run starts the GC loop, running CollectOnce periodically until ctx is cancelled.
// TODO: Implement.
func (gc *GarbageCollector) Run(ctx context.Context) {
	// stub
}
