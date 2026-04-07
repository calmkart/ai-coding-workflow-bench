package gc

import (
	"sync"
	"time"
)

// Store holds resources.
// TODO: Implement Delete (soft delete), Purge (hard delete),
// FindDependents, and other GC-related operations.
type Store struct {
	mu        sync.RWMutex
	resources map[string]*Resource
}

func NewStore() *Store {
	return &Store{resources: make(map[string]*Resource)}
}

func (s *Store) Get(name string) (*Resource, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.resources[name]
	return r, ok
}

func (s *Store) Put(r *Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[r.Metadata.Name] = r
}

// List returns all resources.
func (s *Store) List() []*Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Resource, 0, len(s.resources))
	for _, r := range s.resources {
		result = append(result, r)
	}
	return result
}

// Delete marks a resource for deletion (sets DeletionTimestamp).
// BUG: This should be a soft delete, but currently it's a hard delete.
// No cascade to dependents!
func (s *Store) Delete(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// BUG: Hard delete — should be soft delete with DeletionTimestamp
	delete(s.resources, name)
}

// TODO: Implement Purge (hard delete for resources with DeletionTimestamp and no dependents)
// TODO: Implement FindDependents (find resources owned by given resource)
// TODO: Implement SoftDelete (set DeletionTimestamp)

// softDelete sets the DeletionTimestamp without removing the resource.
func (s *Store) softDelete(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.resources[name]
	if !ok {
		return
	}
	now := time.Now()
	r.Metadata.DeletionTimestamp = &now
}
