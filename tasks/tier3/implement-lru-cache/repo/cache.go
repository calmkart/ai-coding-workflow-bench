package cachebox

import (
	"sync"
)

// SimpleCache is a basic cache with no eviction or TTL.
// PROBLEM: No capacity limit, no LRU eviction, no TTL.
// Need to implement a full LRU cache with:
// - Capacity limit with LRU eviction
// - Per-entry TTL
// - Thread-safe operations
// - O(1) Get/Set via doubly-linked list + map

type SimpleCache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]V
}

func NewSimpleCache[K comparable, V any]() *SimpleCache[K, V] {
	return &SimpleCache[K, V]{
		items: make(map[K]V),
	}
}

func (c *SimpleCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.items[key]
	return val, ok
}

func (c *SimpleCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// PROBLEM: No capacity check, grows unbounded
	c.items[key] = value
}

func (c *SimpleCache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *SimpleCache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
