package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedisClient(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return mr, rdb
}

func TestRedisRateLimiter_UnderLimit(t *testing.T) {
	_, rdb := newTestRedisClient(t)
	rl := NewRedisRateLimiter(rdb, 3, time.Minute, nil)
	defer rl.Stop()
	handler := rl.Handler(okHandler())

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "1.2.3.4:5000"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}
}

func TestRedisRateLimiter_OverLimit(t *testing.T) {
	_, rdb := newTestRedisClient(t)
	rl := NewRedisRateLimiter(rdb, 2, time.Minute, nil)
	defer rl.Stop()
	handler := rl.Handler(okHandler())

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "1.2.3.4:5000"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:5000"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}

func TestRedisRateLimiter_WindowReset(t *testing.T) {
	mr, rdb := newTestRedisClient(t)
	// miniredis minimum TTL resolution is 1s; use 1s window.
	window := 1 * time.Second
	rl := NewRedisRateLimiter(rdb, 1, window, nil)
	defer rl.Stop()
	handler := rl.Handler(okHandler())

	// Exhaust the single token.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:5000"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", rec.Code)
	}

	// Should be limited now.
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:5000"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("limited: expected 429, got %d", rec.Code)
	}

	// Fast-forward miniredis time so the key expires.
	mr.FastForward(window + 100*time.Millisecond)

	// Should pass again.
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:5000"
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("after window reset: expected 200, got %d", rec.Code)
	}
}

func TestRedisRateLimiter_TwoIPsIsolated(t *testing.T) {
	_, rdb := newTestRedisClient(t)
	rl := NewRedisRateLimiter(rdb, 1, time.Minute, nil)
	defer rl.Stop()
	handler := rl.Handler(okHandler())

	// Exhaust IP A.
	for i := 0; i < 1; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "1.1.1.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// IP B should still be fine.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "2.2.2.2:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("IP B should not be rate limited, got %d", rec.Code)
	}
}

func TestRedisRateLimiter_RedisDownFailsOpen(t *testing.T) {
	mr, rdb := newTestRedisClient(t)
	rl := NewRedisRateLimiter(rdb, 1, time.Minute, nil)
	defer rl.Stop()
	handler := rl.Handler(okHandler())

	// Shut down Redis.
	mr.Close()

	// Requests should fail open (200, not 429 or 500).
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:5000"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("redis down: expected fail-open 200, got %d", rec.Code)
	}
}

// Ensure *RedisRateLimiter satisfies the Limiter interface.
var _ Limiter = (*RedisRateLimiter)(nil)

// Ensure *RateLimiter satisfies the Limiter interface.
var _ Limiter = (*RateLimiter)(nil)
