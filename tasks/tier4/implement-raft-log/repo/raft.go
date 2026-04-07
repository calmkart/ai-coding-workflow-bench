package raftlog

import "errors"

// Role represents the node's role in the Raft cluster.
type Role int

const (
	Follower Role = iota
	Leader
)

// LogEntry represents a single entry in the Raft log.
type LogEntry struct {
	Index   int    // 1-based index
	Term    int    // term when entry was created
	Command string // the command to apply
}

// RaftNode implements a simplified Raft node with log replication.
// TODO: Implement AppendEntries, Commit, Apply, Propose.
type RaftNode struct {
	ID          string
	role        Role
	currentTerm int
	log         []LogEntry
	commitIndex int
	lastApplied int
}

var (
	ErrNotLeader    = errors.New("not leader")
	ErrInvalidIndex = errors.New("invalid index")
)

// NewRaftNode creates a new Raft node as a Follower.
func NewRaftNode(id string) *RaftNode {
	return &RaftNode{
		ID:   id,
		role: Follower,
	}
}

// BecomeLeader promotes this node to Leader for the given term.
func (n *RaftNode) BecomeLeader(term int) {
	n.role = Leader
	n.currentTerm = term
}

// Role returns the current role.
func (n *RaftNode) Role() Role {
	return n.role
}

// CurrentTerm returns the current term.
func (n *RaftNode) CurrentTerm() int {
	return n.currentTerm
}

// Propose adds a new log entry (leader only).
// TODO: Implement. Returns error if not leader.
func (n *RaftNode) Propose(command string) (LogEntry, error) {
	return LogEntry{}, ErrNotLeader
}

// AppendEntries implements the Raft AppendEntries RPC.
// Returns (success, currentTerm).
// TODO: Implement the full AppendEntries logic:
// 1. Reject if leaderTerm < currentTerm
// 2. Reject if log doesn't contain entry at prevIndex with prevTerm
// 3. Delete conflicting entries
// 4. Append new entries
// 5. Update currentTerm
func (n *RaftNode) AppendEntries(entries []LogEntry, leaderTerm int, prevIndex int, prevTerm int) (bool, int) {
	return false, n.currentTerm
}

// Commit sets the commit index.
// TODO: Validate that index is within log bounds.
func (n *RaftNode) Commit(index int) error {
	return ErrInvalidIndex
}

// Apply returns committed but not-yet-applied entries and advances lastApplied.
// TODO: Return entries from lastApplied+1 to commitIndex.
func (n *RaftNode) Apply() []LogEntry {
	return nil
}

// GetLog returns the full log.
func (n *RaftNode) GetLog() []LogEntry {
	result := make([]LogEntry, len(n.log))
	copy(result, n.log)
	return result
}

// GetEntry returns the entry at the given 1-based index.
func (n *RaftNode) GetEntry(index int) (LogEntry, bool) {
	if index < 1 || index > len(n.log) {
		return LogEntry{}, false
	}
	return n.log[index-1], true
}

// LastIndex returns the index of the last log entry (0 if empty).
func (n *RaftNode) LastIndex() int {
	return len(n.log)
}

// LastTerm returns the term of the last log entry (0 if empty).
func (n *RaftNode) LastTerm() int {
	if len(n.log) == 0 {
		return 0
	}
	return n.log[len(n.log)-1].Term
}
