package btree

// KV represents a key-value pair.
type KV[K, V any] struct {
	Key   K
	Value V
}

// node is a B-tree node.
type node[K, V any] struct {
	keys     []KV[K, V]
	children []*node[K, V]
	leaf     bool
}

// BTree is a generic B-tree data structure.
// TODO: Implement Insert, Search, Delete, Range, split, merge, borrow.
type BTree[K, V any] struct {
	root  *node[K, V]
	order int
	less  func(K, K) bool
	size  int
}

// NewBTree creates a new B-tree with the given order and comparison function.
// Order must be >= 3.
func NewBTree[K, V any](order int, less func(K, K) bool) *BTree[K, V] {
	if order < 3 {
		panic("btree order must be >= 3")
	}
	return &BTree[K, V]{
		root: &node[K, V]{leaf: true},
		order: order,
		less:  less,
	}
}

// Len returns the number of elements in the tree.
func (t *BTree[K, V]) Len() int {
	return t.size
}

// Insert adds a key-value pair. If the key exists, its value is updated.
// TODO: Implement with node splitting.
func (t *BTree[K, V]) Insert(key K, value V) {
	// stub
}

// Search looks up a key and returns (value, true) or (zero, false).
// TODO: Implement with binary search within nodes.
func (t *BTree[K, V]) Search(key K) (V, bool) {
	var zero V
	return zero, false
}

// Delete removes a key. Returns true if the key existed.
// TODO: Implement with merge/borrow.
func (t *BTree[K, V]) Delete(key K) bool {
	return false
}

// Range returns all key-value pairs with keys in [from, to], sorted.
// TODO: Implement.
func (t *BTree[K, V]) Range(from, to K) []KV[K, V] {
	return nil
}

// Min returns the minimum key-value pair.
// TODO: Implement.
func (t *BTree[K, V]) Min() (KV[K, V], bool) {
	var zero KV[K, V]
	return zero, false
}

// Max returns the maximum key-value pair.
// TODO: Implement.
func (t *BTree[K, V]) Max() (KV[K, V], bool) {
	var zero KV[K, V]
	return zero, false
}

// InOrder returns all key-value pairs in ascending order.
// TODO: Implement.
func (t *BTree[K, V]) InOrder() []KV[K, V] {
	return nil
}
