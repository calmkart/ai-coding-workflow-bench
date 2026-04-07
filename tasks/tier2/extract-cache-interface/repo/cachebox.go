package cachebox

import "sync"

// MapCache is a simple thread-safe cache using a map.
// BUG: No interface - tightly coupled to concrete type.
type MapCache[K comparable, V any] struct {
	mu    sync.RWMutex
	items map[K]V
}

// NewMapCache creates a new MapCache.
// BUG: Returns concrete type instead of interface.
func NewMapCache[K comparable, V any]() *MapCache[K, V] {
	return &MapCache[K, V]{
		items: make(map[K]V),
	}
}

func (c *MapCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.items[key]
	return val, ok
}

func (c *MapCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

func (c *MapCache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *MapCache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
