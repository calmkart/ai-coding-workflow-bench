package hashring

import (
	"sync"
)

// HashRing implements consistent hashing with virtual nodes.
// TODO: Implement all methods.
type HashRing struct {
	mu       sync.RWMutex
	replicas int            // virtual nodes per physical node
	ring     []uint32       // sorted hash values
	nodeMap  map[uint32]string // hash -> physical node name
	nodes    map[string]bool   // set of physical node names
}

// NewHashRing creates a consistent hash ring.
// replicas is the number of virtual nodes per physical node (higher = more uniform distribution).
func NewHashRing(replicas int) *HashRing {
	if replicas < 1 {
		replicas = 1
	}
	return &HashRing{
		replicas: replicas,
		nodeMap:  make(map[uint32]string),
		nodes:    make(map[string]bool),
	}
}

// AddNode adds a physical node with its virtual nodes to the ring.
// TODO: Hash the node name with replica index, add to sorted ring.
func (r *HashRing) AddNode(node string) {
	// stub
}

// RemoveNode removes a physical node and all its virtual nodes.
// TODO: Remove all virtual nodes for this physical node.
func (r *HashRing) RemoveNode(node string) {
	// stub
}

// GetNode returns the node responsible for the given key.
// TODO: Hash the key, find next node on ring clockwise.
func (r *HashRing) GetNode(key string) string {
	return ""
}

// GetNodes returns up to count distinct physical nodes for the key.
// Used for replication — each replica goes to a different physical node.
// TODO: Walk the ring clockwise collecting distinct physical nodes.
func (r *HashRing) GetNodes(key string, count int) []string {
	return nil
}

// NodeCount returns the number of physical nodes.
func (r *HashRing) NodeCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.nodes)
}
