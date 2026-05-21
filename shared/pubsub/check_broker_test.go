package pubsub

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// checkBrokerFactories returns both in-memory and Redis-backed CheckBroker factories.
type checkBrokerFactory struct {
	name string
	new  func(t *testing.T) CheckBroker
}

func checkBrokerFactories(t *testing.T) []checkBrokerFactory {
	t.Helper()
	return []checkBrokerFactory{
		{
			name: "memory",
			new: func(t *testing.T) CheckBroker {
				t.Helper()
				return NewCheckBroker()
			},
		},
		{
			name: "redis",
			new: func(t *testing.T) CheckBroker {
				t.Helper()
				mr := miniredis.RunT(t)
				rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
				t.Cleanup(func() { rdb.Close() })
				ctx, cancel := context.WithCancel(context.Background())
				t.Cleanup(cancel)
				return NewRedisCheckBroker(ctx, rdb)
			},
		},
	}
}

// TestCheckBroker_SubscribeBeforePublish verifies that a subscriber registered
// before any publish receives the event once it is published.
func TestCheckBroker_SubscribeBeforePublish(t *testing.T) {
	for _, f := range checkBrokerFactories(t) {
		f := f
		t.Run(f.name, func(t *testing.T) {
			b := f.new(t)
			ch, unsubscribe := b.Subscribe("page-1")
			defer unsubscribe()

			payload := []byte(`{"status":"success"}`)
			b.Publish("page-1", payload)

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

// TestCheckBroker_SubscribeAfterPublish_CacheHit verifies that a late subscriber
// (connects after publish) receives the cached payload immediately AND remains
// registered as a listener for future events (persistent subscription).
func TestCheckBroker_SubscribeAfterPublish_CacheHit(t *testing.T) {
	for _, f := range checkBrokerFactories(t) {
		f := f
		t.Run(f.name, func(t *testing.T) {
			b := f.new(t)
			payload := []byte(`{"status":"pending"}`)
			b.Publish("page-2", payload)

			// Late subscribe — should get cached payload immediately.
			ch, unsubscribe := b.Subscribe("page-2")
			defer unsubscribe()

			gotCacheHit := false
			timeout := time.After(200 * time.Millisecond)
			for !gotCacheHit {
				select {
				case got := <-ch:
					if string(got) == string(payload) {
						gotCacheHit = true
					}
				case <-timeout:
					t.Error("timed out waiting for cached payload on late subscribe")
					return
				}
			}

			// Drain any duplicate delivery (e.g., Redis pubsub + cache GET both fire).
			// This is expected on the Redis path — both tracks deliver the same payload.
			drainTimeout := time.After(50 * time.Millisecond)
		drain:
			for {
				select {
				case <-ch:
					// Discard any extra copies of the cached payload.
				case <-drainTimeout:
					break drain
				}
			}

			// Listener remains registered — future publishes also arrive.
			payload2 := []byte(`{"status":"complete"}`)
			b.Publish("page-2", payload2)

			select {
			case got := <-ch:
				if string(got) != string(payload2) {
					t.Errorf("future event: got %q, want %q", got, payload2)
				}
			case <-time.After(200 * time.Millisecond):
				t.Error("timed out waiting for future published event (listener should remain)")
			}
		})
	}
}

// TestCheckBroker_NonBlockingPublish verifies that publishing to a full subscriber
// channel does not block the caller (best-effort delivery).
func TestCheckBroker_NonBlockingPublish(t *testing.T) {
	for _, f := range checkBrokerFactories(t) {
		f := f
		t.Run(f.name, func(t *testing.T) {
			b := f.new(t)
			ch, unsubscribe := b.Subscribe("page-3")
			defer unsubscribe()

			// Fill the channel buffer (size 2) without reading.
			b.Publish("page-3", []byte(`{"status":"a"}`))
			b.Publish("page-3", []byte(`{"status":"b"}`))

			// A third publish to a full buffer must not block.
			done := make(chan struct{})
			go func() {
				b.Publish("page-3", []byte(`{"status":"c"}`))
				close(done)
			}()

			select {
			case <-done:
				// Non-blocking: good.
			case <-time.After(200 * time.Millisecond):
				t.Error("Publish blocked on full channel")
			}
			_ = ch
		})
	}
}

// TestCheckBroker_MultiSubscriberFanout verifies that all active subscribers
// receive the published payload.
func TestCheckBroker_MultiSubscriberFanout(t *testing.T) {
	for _, f := range checkBrokerFactories(t) {
		f := f
		t.Run(f.name, func(t *testing.T) {
			b := f.new(t)

			ch1, unsub1 := b.Subscribe("page-fan")
			defer unsub1()
			ch2, unsub2 := b.Subscribe("page-fan")
			defer unsub2()

			payload := []byte(`{"status":"fan-out"}`)
			b.Publish("page-fan", payload)

			for i, ch := range []<-chan []byte{ch1, ch2} {
				select {
				case got := <-ch:
					if string(got) != string(payload) {
						t.Errorf("subscriber %d: got %q, want %q", i+1, got, payload)
					}
				case <-time.After(200 * time.Millisecond):
					t.Errorf("subscriber %d: timed out", i+1)
				}
			}
		})
	}
}

// TestCheckBroker_CacheTTLExpiry verifies that a subscriber connecting after
// the cache TTL has expired receives no replay payload.
// We only test this for the Redis backend (miniredis FastForward is exact);
// the memory backend's evict loop runs on an interval and is impractical to
// test without sleeping. The behavioral contract is the same.
func TestCheckBroker_CacheTTLExpiry(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b := NewRedisCheckBroker(ctx, rdb)

	b.Publish("page-ttl", []byte(`{"status":"ttl-test"}`))

	// Give the background goroutine time to drain the pubsub message so it
	// doesn't deliver it to future listeners after FastForward.
	time.Sleep(50 * time.Millisecond)

	// Advance past the 2-minute cache TTL — the Redis cache key is now expired.
	mr.FastForward(3 * time.Minute)

	ch, unsubscribe := b.Subscribe("page-ttl")
	defer unsubscribe()

	// No cached payload should arrive — channel should be empty after 100ms.
	select {
	case got := <-ch:
		t.Errorf("expected no cached payload after TTL expiry, got %q", got)
	case <-time.After(100 * time.Millisecond):
		// Correct — nothing in channel.
	}
}

// TestRedisCheckBroker_MultiInstanceFanout verifies that two instances sharing
// the same Redis both receive events published by either one.
func TestRedisCheckBroker_MultiInstanceFanout(t *testing.T) {
	mr := miniredis.RunT(t)

	rdbA := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rdbB := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdbA.Close()
	defer rdbB.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	brokerA := NewRedisCheckBroker(ctx, rdbA)
	brokerB := NewRedisCheckBroker(ctx, rdbB)

	// Instance B subscribes.
	chB, unsubB := brokerB.Subscribe("page-multi")
	defer unsubB()

	// Wait briefly for the Redis PSUBSCRIBE to be established.
	time.Sleep(50 * time.Millisecond)

	// Instance A publishes.
	payload := []byte(`{"status":"cross-instance"}`)
	brokerA.Publish("page-multi", payload)

	select {
	case got := <-chB:
		if string(got) != string(payload) {
			t.Errorf("got %q, want %q", got, payload)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("timed out waiting for cross-instance event on broker B")
	}
}
