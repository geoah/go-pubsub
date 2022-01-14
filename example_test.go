package pubsub

import (
	"fmt"
	"testing"
)

func Test_PubSub_Example_Basic(t *testing.T) {
	psString := NewPubSub[string]()
	subString := psString.Subscribe("foo")
	psString.Publish("foo", "bar")
	fmt.Println("<<", <-subString)
	psString.Publish("foo", "not-bar")
	fmt.Println("<<", <-subString)
}

func Test_PubSub_Example_Struct(t *testing.T) {
	type Message struct {
		Something string
	}

	psStruct := NewPubSub[Message]()
	subStruct := psStruct.Subscribe("foo")
	psStruct.Publish("foo", Message{"bar"})
	fmt.Println("<<", <-subStruct)
	psStruct.Publish("foo", Message{"not-bar"})
	fmt.Println("<<", <-subStruct)
}
