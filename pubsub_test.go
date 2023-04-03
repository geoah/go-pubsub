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

func Test_PubSub_Filters(t *testing.T) {
	ps := NewPubSub[string]()

	filter1checks := 0
	filter1hits := 0
	filter2checks := 0
	filter2hits := 0

	sub1 := ps.Subscribe("foo", func(s string) bool {
		filter1checks++
		if s == "f01" {
			filter1hits++
			return true
		}
		return false
	})

	sub2 := ps.Subscribe("foo", func(s string) bool {
		filter2checks++
		if s == "f02" {
			filter2hits++
			return true
		}
		return false
	})

	t.Run("two subscriptions, ok", func(t *testing.T) {
		ps.Publish("foo", "f02")
		require.Equal(t, "f02", <-sub2.Channel())
	})

	t.Run("one subscription, ok", func(t *testing.T) {
		ps.Publish("foo", "f01")
		require.Equal(t, "f01", <-sub1.Channel())
	})

	require.Equal(t, 2, filter1checks)
	require.Equal(t, 1, filter1hits)
	require.Equal(t, 2, filter2checks)
	require.Equal(t, 1, filter2hits)
}
