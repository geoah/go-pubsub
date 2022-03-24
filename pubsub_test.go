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
		require.Equal(t, "f01", <-sub1.Channel())
	})

	sub2 := ps.Subscribe("foo")
	t.Run("two subscriptions, ok", func(t *testing.T) {
		ps.Publish("foo", "f02")
		n, err := sub1.Next()
		require.NoError(t, err)
		require.Equal(t, "f02", n)
		n = <-sub2.Channel()
		require.Equal(t, "f02", n)
	})

	t.Run("two subscriptions, one blocking, ok", func(t *testing.T) {
		ps.Publish("foo", "f03")
		require.Equal(t, "f03", <-sub1.Channel())
		ps.Publish("foo", "f04")
		require.Equal(t, "f04", <-sub1.Channel())
	})
}
