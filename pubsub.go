package pubsub

import (
	"fmt"
	"sync"
	"time"
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
func (ps *PubSub[Value]) Subscribe(
	topic string,
	filters ...func(Value) bool,
) *Subscription[Value] {
	t := ps.getTopic(topic)
	return t.Subscribe(filters...)
}

// Publish an event to all a topic's subscribers
func (ps *PubSub[Value]) Publish(topic string, value Value) {
	t := ps.getTopic(topic)
	t.Publish(value)
}

// Topic contains all the events we are subscribed to.
type Topic[Value any] struct {
	lock               sync.RWMutex
	subscriptions      []*Subscription[Value]
	defaultAckDeadline time.Duration
	defaultBufferSize  int
}

type Subscription[Value any] struct {
	ackDeadline time.Duration
	buffer      chan valueExpiry[Value]
	values      chan Value
	cancel      chan<- struct{}
	filters     []func(Value) bool
}

type valueExpiry[Value any] struct {
	value   Value
	created time.Time
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

func (s *Subscription[Value]) publish(value Value) {
	for _, f := range s.filters {
		if !f(value) {
			return
		}
	}
	out := valueExpiry[Value]{
		value:   value,
		created: time.Now(),
	}
	select {
	case s.buffer <- out:
	default:
	}
}

func (s *Subscription[Value]) process(cancel <-chan struct{}) {
	tryPub := func(v valueExpiry[Value]) {
		select {
		case <-cancel:
			close(s.values)
			return
		case s.values <- v.value:
			return
		case <-time.After(time.Until(v.created.Add(s.ackDeadline))):
			return
		}
	}
	for {
		select {
		case <-cancel:
			close(s.values)
			return
		case v := <-s.buffer:
			tryPub(v)
		}
	}
}

// NewTopic creates a new topic.
func NewTopic[Value any]() *Topic[Value] {
	return &Topic[Value]{
		subscriptions:      []*Subscription[Value]{},
		defaultAckDeadline: 0,
		defaultBufferSize:  1,
	}
}

// NewTopicWithOptions creates a new custom topic.
// WARNING: This is an experimental API and WILL change in the future.
func NewTopicWithOptions[Value any](
	ackDeadline time.Duration,
	bufferSize int,
) *Topic[Value] {
	return &Topic[Value]{
		subscriptions:      []*Subscription[Value]{},
		defaultAckDeadline: ackDeadline,
		defaultBufferSize:  bufferSize,
	}
}

// Subscribe to messages published to this topic.
func (t *Topic[Value]) Subscribe(
	filters ...func(Value) bool,
) *Subscription[Value] {
	return t.SubscribeWithOptions(t.defaultBufferSize, t.defaultAckDeadline, filters...)
}

// SubscribeWithOptions subscribes to messages published to this topic with
// custom options.
// WARNING: This is an experimental API and WILL change in the future.
func (t *Topic[Value]) SubscribeWithOptions(
	bufferSize int,
	ackDeadline time.Duration,
	filters ...func(Value) bool,
) *Subscription[Value] {
	t.lock.Lock()
	defer t.lock.Unlock()
	cancel := make(chan struct{})
	s := &Subscription[Value]{
		ackDeadline: ackDeadline,
		buffer:      make(chan valueExpiry[Value], bufferSize),
		values:      make(chan Value),
		cancel:      cancel,
		filters:     filters,
	}
	go s.process(cancel)
	t.subscriptions = append(t.subscriptions, s)
	return s
}

// Publish an event to all the topic's subscribers.
func (t *Topic[Value]) Publish(value Value) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	for _, sub := range t.subscriptions {
		sub.publish(value)
	}
}
