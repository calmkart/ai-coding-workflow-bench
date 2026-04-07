package eventbus

import "sync"

// Event represents an event with a topic and data.
type Event struct {
	Topic string
	Data  any
}

// Handler is a function that handles an event.
type Handler func(Event)

// SubscriptionID uniquely identifies a subscription.
type SubscriptionID int64

// EventBus is a simple event bus.
// TODO: Implement Subscribe, Publish, Unsubscribe, Close.
// Currently this is just a skeleton with no real functionality.
type EventBus struct {
	mu      sync.RWMutex
	nextID  SubscriptionID
	closed  bool
}

// NewEventBus creates a new event bus.
func NewEventBus() *EventBus {
	return &EventBus{}
}

// Subscribe registers a handler for a topic.
// TODO: Implement subscription storage and topic matching.
func (eb *EventBus) Subscribe(topic string, handler Handler) SubscriptionID {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.nextID++
	// PROBLEM: Handler is not stored anywhere
	return eb.nextID
}

// Publish sends an event to all handlers subscribed to the topic.
// TODO: Implement event delivery.
func (eb *EventBus) Publish(topic string, data any) {
	// PROBLEM: No handlers to deliver to
}

// Unsubscribe removes a subscription.
// TODO: Implement removal.
func (eb *EventBus) Unsubscribe(id SubscriptionID) {
	// PROBLEM: Nothing to remove
}

// Close shuts down the event bus.
func (eb *EventBus) Close() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.closed = true
}
