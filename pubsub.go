package pubsub

import (
	"sync"
)

// PubSub provides the ability to publish and subscribe to topics.
// Multiple subscriptions can exist on a single topic, and all will receive the
// same messages.
type PubSub[Value any] struct {
	lock   sync.RWMutex
	topics map[string]*Topic[Value]
}

// NewPubSub creates a new pubsub.
func NewPubSub[Value any]() *PubSub[Value] {
	return &PubSub[Value]{
		topics: map[string]*Topic[Value]{},
	}
}

// getTopic returns an existing or creates a new topic
func (ps *PubSub[Value]) getTopic(topic string) *Topic[Value] {
	ps.lock.RLock()
	if t, ok := ps.topics[topic]; ok {
		ps.lock.RUnlock()
		return t
	}
	ps.lock.RUnlock()
	ps.lock.Lock()
	defer ps.lock.Unlock()
	t := NewTopic[Value]()
	ps.topics[topic] = t
	return t
}

// Subscribe to a topic
func (ps *PubSub[Value]) Subscribe(topic string) <-chan Value {
	t := ps.getTopic(topic)
	return t.Subscribe()
}

// Publish an event to all a topic's subscribers
func (ps *PubSub[Value]) Publish(topic string, value Value) {
	t := ps.getTopic(topic)
	t.Publish(value)
}

// Topic contains all the events we are subscribed to.
type Topic[Value any] struct {
	lock          sync.RWMutex
	subscriptions []chan Value
}

// NewTopic creates a new topic.
func NewTopic[Value any]() *Topic[Value] {
	return &Topic[Value]{
		subscriptions: []chan Value{},
	}
}

// Subscribe to messages published to this topic.
func (t *Topic[Value]) Subscribe() <-chan Value {
	t.lock.Lock()
	defer t.lock.Unlock()
	c := make(chan Value, 1)
	t.subscriptions = append(t.subscriptions, c)
	return c
}

// Publish an event to all the topic's subscribers.
func (t *Topic[Value]) Publish(value Value) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	for _, sub := range t.subscriptions {
		select {
		case sub <- value:
		default:
		}
	}
}
