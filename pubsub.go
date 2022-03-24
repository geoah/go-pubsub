package pubsub

import (
	"fmt"
	"sync"
)

var ErrSubscriptionClosed = fmt.Errorf("subscription closed")

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
func (ps *PubSub[Value]) Subscribe(topic string) *Subscription[Value] {
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

type Subscription[Value any] struct {
	values <-chan Value
	cancel chan<- struct{}
}

func (s *Subscription[Value]) Cancel() {
	close(s.cancel)
}

func (s *Subscription[Value]) Channel() <-chan Value {
	return s.values
}

func (s *Subscription[Value]) Next() (Value, error) {
	v, ok := <-s.values
	if !ok {
		return v, ErrSubscriptionClosed
	}
	return v, nil
}

// NewTopic creates a new topic.
func NewTopic[Value any]() *Topic[Value] {
	return &Topic[Value]{
		subscriptions: []chan Value{},
	}
}

// Subscribe to messages published to this topic.
func (t *Topic[Value]) Subscribe() *Subscription[Value] {
	t.lock.Lock()
	defer t.lock.Unlock()
	values := make(chan Value, 1)
	cancel := make(chan struct{})
	t.subscriptions = append(t.subscriptions, values)
	s := &Subscription[Value]{
		values: values,
		cancel: cancel,
	}
	go func() {
		<-cancel
		close(values)
	}()
	return s
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
