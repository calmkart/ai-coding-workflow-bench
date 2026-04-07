package pubsub

import "sync"

// Message represents a published message.
type Message struct {
	Topic string
	Data  any
}

// PubSub is an in-memory publish-subscribe system.
// TODO: Implement Subscribe, Publish, Unsubscribe, Close.
type PubSub struct {
	mu     sync.RWMutex
	closed bool
}

// NewPubSub creates a new PubSub system.
func NewPubSub() *PubSub {
	return &PubSub{}
}

// Subscribe returns a channel that receives messages for the given topic.
// TODO: Implement subscription storage and channel creation.
func (ps *PubSub) Subscribe(topic string) <-chan Message {
	// PROBLEM: Returns nil channel - subscriber would block forever
	return nil
}

// Publish sends a message to all subscribers of the topic.
// TODO: Implement message delivery.
func (ps *PubSub) Publish(topic string, data any) error {
	// PROBLEM: No subscribers to deliver to
	return nil
}

// Unsubscribe removes a subscription.
// TODO: Implement unsubscription.
func (ps *PubSub) Unsubscribe(topic string, ch <-chan Message) {
	// PROBLEM: Nothing to unsubscribe
}

// Close shuts down the PubSub system.
func (ps *PubSub) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.closed = true
	// PROBLEM: No channels to close
}
