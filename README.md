# go-pubsub

Proof of concept channel-based pub/sub using go1.18 generics.

## Notes

* Subscriptions currently are buffered with a size of 1

## Examples

### Basic types

```go
// create a new pubsub with a message of type string
ps := pubsub.NewPubSub[string]()

// create a subscription to a new topic
sub := ps.Subscribe("foo")

// publish, and then consume the message
ps.Publish("foo", "bar")
fmt.Println("<<", <-sub)
```

### Structs

```go
type Message struct {
	Something string
}

// create a new pubsub with a message of type Message
ps := pubsub.NewPubSub[Message]()

// create a subscription to a new topic
sub := ps.Subscribe("foo")

// publish, and then consume the message
ps.Publish("foo", Message{"bar"})
fmt.Println("<<", <-sub)
```
