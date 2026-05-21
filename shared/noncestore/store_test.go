package noncestore

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// factory returns a fresh NonceStore for each test run.
type factory struct {
	name string
	new  func(t *testing.T) NonceStore
}

func factories(t *testing.T) []factory {
	t.Helper()
	return []factory{
		{
			name: "memory",
			new: func(t *testing.T) NonceStore {
				t.Helper()
				return New()
			},
		},
		{
			name: "redis",
			new: func(t *testing.T) NonceStore {
				t.Helper()
				mr := miniredis.RunT(t)
				rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
				t.Cleanup(func() { rdb.Close() })
				return NewRedisStore(rdb)
			},
		},
	}
}

func TestSaveAndConsume(t *testing.T) {
	for _, f := range factories(t) {
		f := f
		t.Run(f.name, func(t *testing.T) {
			s := f.new(t)
			entry := NonceEntry{
				AccessToken:  "at-123",
				RefreshToken: "rt-456",
				ExpiresIn:    900,
			}
			s.Save("nonce-1", entry)

			got := s.Consume("nonce-1")
			if got == nil {
				t.Fatal("expected entry, got nil")
			}
			if got.AccessToken != "at-123" {
				t.Errorf("AccessToken = %q, want %q", got.AccessToken, "at-123")
			}
			if got.RefreshToken != "rt-456" {
				t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, "rt-456")
			}
			if got.ExpiresIn != 900 {
				t.Errorf("ExpiresIn = %d, want 900", got.ExpiresIn)
			}

			// Second consume should return nil (one-time use).
			if s.Consume("nonce-1") != nil {
				t.Error("expected nil on second consume")
			}
		})
	}
}

func TestPeekDoesNotConsume(t *testing.T) {
	for _, f := range factories(t) {
		f := f
		t.Run(f.name, func(t *testing.T) {
			s := f.new(t)
			s.Save("nonce-2", NonceEntry{AccessToken: "at-peek"})

			got := s.Peek("nonce-2")
			if got == nil || got.AccessToken != "at-peek" {
				t.Fatal("Peek should return the entry")
			}

			// Peek again — should still be there.
			got = s.Peek("nonce-2")
			if got == nil || got.AccessToken != "at-peek" {
				t.Fatal("Peek should not consume the entry")
			}

			// Consume should still work after multiple peeks.
			got = s.Consume("nonce-2")
			if got == nil || got.AccessToken != "at-peek" {
				t.Fatal("Consume should return the entry after peeks")
			}
		})
	}
}

func TestConsumeNonExistent(t *testing.T) {
	for _, f := range factories(t) {
		f := f
		t.Run(f.name, func(t *testing.T) {
			s := f.new(t)
			if s.Consume("does-not-exist") != nil {
				t.Error("expected nil for non-existent nonce")
			}
		})
	}
}

func TestPeekNonExistent(t *testing.T) {
	for _, f := range factories(t) {
		f := f
		t.Run(f.name, func(t *testing.T) {
			s := f.new(t)
			if s.Peek("does-not-exist") != nil {
				t.Error("expected nil for non-existent nonce")
			}
		})
	}
}

// TestConsumeExpiredNonce verifies that expired entries return nil.
// For MemoryStore, we manipulate internal state; for RedisStore we use
// miniredis FastForward to advance the clock.
func TestConsumeExpiredNonce(t *testing.T) {
	// Memory variant — inject pre-expired entry.
	t.Run("memory", func(t *testing.T) {
		s := New()
		entry := NonceEntry{
			AccessToken:  "expired-token",
			RefreshToken: "expired-refresh",
			ExpiresIn:    900,
			CreatedAt:    time.Now().Add(-31 * time.Second), // already expired
		}
		s.mu.Lock()
		s.entries["expired-nonce"] = entry
		s.mu.Unlock()

		if s.Consume("expired-nonce") != nil {
			t.Error("expected nil for expired nonce on Consume")
		}

		// Re-insert for Peek test.
		s.mu.Lock()
		s.entries["expired-nonce-2"] = entry
		s.mu.Unlock()

		if s.Peek("expired-nonce-2") != nil {
			t.Error("expected nil for expired nonce on Peek")
		}
	})

	// Redis variant — use miniredis FastForward.
	t.Run("redis", func(t *testing.T) {
		mr := miniredis.RunT(t)
		rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		defer rdb.Close()
		s := NewRedisStore(rdb)

		s.Save("exp-redis", NonceEntry{AccessToken: "will-expire"})
		// Advance miniredis clock past the 30s TTL.
		mr.FastForward(31 * time.Second)

		if s.Consume("exp-redis") != nil {
			t.Error("expected nil for expired nonce on Consume (redis)")
		}

		// A fresh entry to test Peek expiry.
		s.Save("exp-redis-2", NonceEntry{AccessToken: "will-expire-2"})
		mr.FastForward(31 * time.Second)

		if s.Peek("exp-redis-2") != nil {
			t.Error("expected nil for expired nonce on Peek (redis)")
		}
	})
}

func TestSavePrunesExpired(t *testing.T) {
	// Only makes sense for MemoryStore (Redis uses TTL natively).
	t.Run("memory", func(t *testing.T) {
		s := New()

		// Insert an already-expired entry.
		s.mu.Lock()
		s.entries["old"] = NonceEntry{
			AccessToken: "old-token",
			CreatedAt:   time.Now().Add(-31 * time.Second),
		}
		s.mu.Unlock()

		// Save a new entry — should prune the old one.
		s.Save("new", NonceEntry{AccessToken: "new-token"})

		s.mu.Lock()
		_, oldExists := s.entries["old"]
		s.mu.Unlock()

		if oldExists {
			t.Error("expected expired entry to be pruned on Save")
		}
	})
}

// TestMultiInstanceCrossNodePeek simulates two instances sharing Redis:
// instance A saves, instance B peeks. Uses two *RedisStore backed by the same miniredis.
func TestMultiInstanceCrossNodePeek(t *testing.T) {
	mr := miniredis.RunT(t)

	rdbA := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rdbB := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdbA.Close()
	defer rdbB.Close()

	instanceA := NewRedisStore(rdbA)
	instanceB := NewRedisStore(rdbB)

	instanceA.Save("cross-nonce", NonceEntry{AccessToken: "shared-token", RefreshToken: "shared-refresh", ExpiresIn: 300})

	got := instanceB.Peek("cross-nonce")
	if got == nil {
		t.Fatal("instance B should see entry saved by instance A")
	}
	if got.AccessToken != "shared-token" {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, "shared-token")
	}

	// Consume on B — A's token is gone.
	consumed := instanceB.Consume("cross-nonce")
	if consumed == nil {
		t.Fatal("Consume on instance B should succeed")
	}
	if instanceA.Peek("cross-nonce") != nil {
		t.Error("after Consume on B, Peek on A should return nil")
	}
}

// TestAtomicConsumeUnderConcurrency verifies that exactly one goroutine receives
// the nonce when many goroutines race to consume it.
func TestAtomicConsumeUnderConcurrency(t *testing.T) {
	for _, f := range factories(t) {
		f := f
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()
			s := f.new(t)
			s.Save("race-nonce", NonceEntry{AccessToken: "contested"})

			const goroutines = 20
			results := make([]*NonceEntry, goroutines)
			var wg sync.WaitGroup
			wg.Add(goroutines)
			for i := 0; i < goroutines; i++ {
				i := i
				go func() {
					defer wg.Done()
					results[i] = s.Consume("race-nonce")
				}()
			}
			wg.Wait()

			successCount := 0
			for _, r := range results {
				if r != nil {
					successCount++
				}
			}
			if successCount != 1 {
				t.Errorf("exactly 1 goroutine should succeed, got %d", successCount)
			}
		})
	}
}

// TestRedisStoreBackedByClient verifies the RedisStore properly talks to Redis.
func TestRedisStoreRoundTrip(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	s := NewRedisStore(rdb)
	s.Save("rt-nonce", NonceEntry{AccessToken: "token-rt", ExpiresIn: 42})

	// Verify key exists in miniredis.
	ctx := context.Background()
	ttl, err := rdb.TTL(ctx, "nonce:rt-nonce").Result()
	if err != nil {
		t.Fatalf("TTL error: %v", err)
	}
	if ttl <= 0 || ttl > 30*time.Second {
		t.Errorf("expected TTL between 0-30s, got %v", ttl)
	}

	// Peek should work without consuming.
	got := s.Peek("rt-nonce")
	if got == nil || got.AccessToken != "token-rt" {
		t.Errorf("Peek returned %v, want {AccessToken: token-rt}", got)
	}

	// Key should still exist in Redis after Peek.
	exists, _ := rdb.Exists(ctx, "nonce:rt-nonce").Result()
	if exists != 1 {
		t.Error("key should still exist after Peek")
	}

	// Consume removes it.
	consumed := s.Consume("rt-nonce")
	if consumed == nil || consumed.AccessToken != "token-rt" {
		t.Error("Consume should return the entry")
	}
	exists, _ = rdb.Exists(ctx, "nonce:rt-nonce").Result()
	if exists != 0 {
		t.Error("key should be gone after Consume")
	}
}
