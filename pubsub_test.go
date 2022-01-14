package pubsub

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_PubSub(t *testing.T) {
	ps := NewPubSub[string]()

	sub1 := ps.Subscribe("foo")
	t.Run("one subscription, ok", func(t *testing.T) {
		ps.Publish("foo", "f01")
		require.Equal(t, "f01", <-sub1)
	})

	sub2 := ps.Subscribe("foo")
	t.Run("two subscriptions, ok", func(t *testing.T) {
		ps.Publish("foo", "f02")
		require.Equal(t, "f02", <-sub1)
		require.Equal(t, "f02", <-sub2)
	})

	t.Run("two subscriptions, one blocking, ok", func(t *testing.T) {
		ps.Publish("foo", "f03")
		require.Equal(t, "f03", <-sub1)
		ps.Publish("foo", "f04")
		require.Equal(t, "f04", <-sub1)
	})
}
