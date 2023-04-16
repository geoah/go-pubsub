package pubsub

import (
	"testing"
	"time"

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

func Test_PubSub_Deadline(t *testing.T) {
	t1 := NewTopicWithOptions[string](time.Millisecond*250, 10)

	s1 := t1.Subscribe()
	s2 := t1.Subscribe()

	go func() {
		time.Sleep(time.Second)
		s1.Cancel()
		s2.Cancel()
	}()

	t.Run("try 1: only s1 acks, ok", func(t *testing.T) {
		t1.Publish("f01")
		require.Equal(t, "f01", <-s1.Channel())
	})

	t.Run("try 2 (after 100ms): only s1 acks, ok", func(t *testing.T) {
		time.Sleep(time.Millisecond * 100)
		t1.Publish("f02")
		require.Equal(t, "f02", <-s1.Channel())
	})

	t.Run("try 3 (after 250ms): both ack, ok", func(t *testing.T) {
		time.Sleep(time.Millisecond * 150)
		t1.Publish("f03")
		require.Equal(t, "f03", <-s1.Channel())
		require.Equal(t, "f02", <-s2.Channel())
		require.Equal(t, "f03", <-s2.Channel())
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
