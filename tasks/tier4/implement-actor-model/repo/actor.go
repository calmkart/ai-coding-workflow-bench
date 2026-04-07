package actor

// Message represents a message sent between actors.
type Message struct {
	Type    string
	Payload any
	From    *ActorRef // sender reference for Reply
}

// ActorContext provides context for message processing.
type ActorContext interface {
	Self() *ActorRef
	System() *ActorSystem
	Reply(msg Message)
}

// Actor processes messages.
type Actor interface {
	Receive(ctx ActorContext, msg Message)
}

// ActorRef is a reference to a running actor, used to send messages.
// TODO: Implement Send.
type ActorRef struct {
	name    string
	mailbox chan Message
	system  *ActorSystem
}

// Name returns the actor's name.
func (ref *ActorRef) Name() string {
	return ref.name
}

// Send delivers a message to this actor's mailbox.
// TODO: Implement.
func (ref *ActorRef) Send(msg Message) {
	// stub
}

// Mailbox represents a message queue for an actor.
type Mailbox struct {
	ch chan Message
}

// ActorSystem manages actor lifecycles.
// TODO: Implement Spawn, Shutdown, Lookup.
type ActorSystem struct {
	actors map[string]*ActorRef
	done   chan struct{}
}

// NewActorSystem creates a new actor system.
func NewActorSystem() *ActorSystem {
	return &ActorSystem{
		actors: make(map[string]*ActorRef),
		done:   make(chan struct{}),
	}
}

// Spawn creates and starts a new actor.
// TODO: Create mailbox, start goroutine to process messages.
func (s *ActorSystem) Spawn(name string, actor Actor) *ActorRef {
	return nil
}

// Lookup finds an actor by name.
// TODO: Implement.
func (s *ActorSystem) Lookup(name string) *ActorRef {
	return nil
}

// Shutdown gracefully stops all actors.
// TODO: Close all mailboxes and wait for goroutines to finish.
func (s *ActorSystem) Shutdown() {
	// stub
}
