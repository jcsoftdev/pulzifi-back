package pubsub

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type insightBrokerFactory struct {
	name string
	new  func(t *testing.T) InsightBroker
}

func insightBrokerFactories(t *testing.T) []insightBrokerFactory {
	t.Helper()
	return []insightBrokerFactory{
		{
			name: "memory",
			new: func(t *testing.T) InsightBroker {
				t.Helper()
				return NewInsightBroker()
			},
		},
		{
			name: "redis",
			new: func(t *testing.T) InsightBroker {
				t.Helper()
				mr := miniredis.RunT(t)
				rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
				t.Cleanup(func() { rdb.Close() })
				ctx, cancel := context.WithCancel(context.Background())
				t.Cleanup(cancel)
				return NewRedisInsightBroker(ctx, rdb)
			},
		},
	}
}

// TestInsightBroker_SubscribeBeforePublish verifies that a subscriber registered
// before any publish receives the event once, then the channel receives no further
// messages (one-shot — the unsubscribe func closes gracefully).
func TestInsightBroker_SubscribeBeforePublish(t *testing.T) {
	for _, f := range insightBrokerFactories(t) {
		f := f
		t.Run(f.name, func(t *testing.T) {
			b := f.new(t)
			ch, unsubscribe := b.Subscribe("check-1")
			defer unsubscribe()

			payload := []byte(`{"type":"marketing","content":"test insight"}`)
			b.Publish("check-1", payload)

			select {
			case got := <-ch:
				if string(got) != string(payload) {
					t.Errorf("got %q, want %q", got, payload)
				}
			case <-time.After(200 * time.Millisecond):
				t.Error("timed out waiting for published payload")
			}
		})
	}
}

// TestInsightBroker_SubscribeAfterPublish_CacheHit verifies one-shot cache hit semantics:
// - Late subscriber receives cached payload immediately.
// - Channel is closed (or no further events arrive) after the one-shot delivery.
// - No persistent listener is registered (a second publish is NOT delivered).
func TestInsightBroker_SubscribeAfterPublish_OneShot(t *testing.T) {
	for _, f := range insightBrokerFactories(t) {
		f := f
		t.Run(f.name, func(t *testing.T) {
			b := f.new(t)
			payload := []byte(`{"type":"analysis","content":"done"}`)
			b.Publish("check-2", payload)

			// Late subscribe — one-shot delivery expected.
			ch, unsubscribe := b.Subscribe("check-2")
			defer unsubscribe()

			select {
			case got := <-ch:
				if string(got) != string(payload) {
					t.Errorf("one-shot cache hit: got %q, want %q", got, payload)
				}
			case <-time.After(200 * time.Millisecond):
				t.Error("timed out waiting for one-shot cache delivery")
				return
			}

			// Verify no additional payload arrives (channel is quiet after one-shot).
			// A second publish should NOT be delivered because no listener was registered.
			b.Publish("check-2", []byte(`{"type":"unexpected","content":"should not arrive"}`))
			select {
			case unexpected := <-ch:
				t.Errorf("expected no second delivery after one-shot cache hit, got %q", unexpected)
			case <-time.After(100 * time.Millisecond):
				// Correct — channel is quiet.
			}
		})
	}
}

// TestInsightBroker_ChannelClosedAfterDelivery verifies that on cache hit the
// channel is closed (or returns the value and then blocks) after delivery.
// For the memory backend the channel is NOT closed on cache hit (it's buffered(1)).
// For the Redis backend the behavior should also be: deliver once, no lingering subscription.
func TestInsightBroker_NoLingeringSubscription(t *testing.T) {
	for _, f := range insightBrokerFactories(t) {
		f := f
		t.Run(f.name, func(t *testing.T) {
			b := f.new(t)
			b.Publish("check-3", []byte(`{"ready":true}`))

			ch, unsubscribe := b.Subscribe("check-3")
			defer unsubscribe()

			// Drain the one-shot delivery.
			select {
			case <-ch:
			case <-time.After(200 * time.Millisecond):
				t.Error("did not receive one-shot delivery")
				return
			}

			// No further events should arrive — the listener is not registered.
			b.Publish("check-3", []byte(`{"ready":"second"}`))
			select {
			case msg := <-ch:
				t.Errorf("unexpected second message after one-shot: %q", msg)
			case <-time.After(100 * time.Millisecond):
				// Correct.
			}
		})
	}
}

// TestRedisInsightBroker_MultiInstanceCacheHit verifies that a late subscriber
// on instance B receives a cached payload published by instance A.
func TestRedisInsightBroker_MultiInstanceCacheHit(t *testing.T) {
	mr := miniredis.RunT(t)
	rdbA := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rdbB := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdbA.Close()
	defer rdbB.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	brokerA := NewRedisInsightBroker(ctx, rdbA)
	brokerB := NewRedisInsightBroker(ctx, rdbB)

	payload := []byte(`{"type":"marketing","result":"cross-instance"}`)
	brokerA.Publish("cross-check", payload)

	// Instance B subscribes late — should get cached payload.
	ch, unsubscribe := brokerB.Subscribe("cross-check")
	defer unsubscribe()

	select {
	case got := <-ch:
		if string(got) != string(payload) {
			t.Errorf("got %q, want %q", got, payload)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("instance B timed out waiting for cached payload from instance A")
	}
}

// TestInsightBroker_NonBlockingPublish verifies that publishing does not block.
func TestInsightBroker_NonBlockingPublish(t *testing.T) {
	for _, f := range insightBrokerFactories(t) {
		f := f
		t.Run(f.name, func(t *testing.T) {
			b := f.new(t)
			ch, unsubscribe := b.Subscribe("check-nb")
			defer unsubscribe()

			// Fill the buffer without reading.
			b.Publish("check-nb", []byte(`{"a":1}`))

			// Second publish must not block even if channel is full.
			done := make(chan struct{})
			go func() {
				b.Publish("check-nb", []byte(`{"b":2}`))
				close(done)
			}()

			select {
			case <-done:
			case <-time.After(200 * time.Millisecond):
				t.Error("Publish blocked on full channel")
			}
			_ = ch
		})
	}
}
