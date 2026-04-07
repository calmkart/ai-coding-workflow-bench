package raftlog

import "testing"

func TestNewRaftNode(t *testing.T) {
	node := NewRaftNode("node-1")
	if node.Role() != Follower {
		t.Fatal("new node should be Follower")
	}
	if node.LastIndex() != 0 {
		t.Fatal("new node should have empty log")
	}
}
