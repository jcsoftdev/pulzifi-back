package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Limiter is the common interface satisfied by both the in-memory RateLimiter
// and the Redis-backed RedisRateLimiter.  Use this type in wiring code so the
// backend can be swapped without changing callers.
type Limiter interface {
	Handler(next http.Handler) http.Handler
	Stop()
}

// RedisRateLimiter is a fixed-window rate limiter backed by Redis.
//
// Key format: ratelimit:<clientIP>:<window-bucket>
//
// Algorithm (pipeline INCR + conditional EXPIRE):
//  1. INCR the key — if result == 1 the key is new, set EXPIRE = window.
//  2. If result > max → 429.
//
// On any Redis error the request is allowed through (fail-open) so that a
// transient Redis blip does not take down the API.
type RedisRateLimiter struct {
	rdb          *redis.Client
	max          int
	window       time.Duration
	trustedCIDRs []net.IPNet
}

// NewRedisRateLimiter creates a Redis-backed rate limiter.
// trusted is passed to ClientIP for real-IP extraction; pass nil to use RemoteAddr directly.
func NewRedisRateLimiter(rdb *redis.Client, max int, window time.Duration, trusted []net.IPNet) *RedisRateLimiter {
	return &RedisRateLimiter{
		rdb:          rdb,
		max:          max,
		window:       window,
		trustedCIDRs: trusted,
	}
}

// Stop is a no-op for RedisRateLimiter (no background goroutines).
// Implemented to satisfy the Limiter interface.
func (rl *RedisRateLimiter) Stop() {}

// Handler returns a middleware that enforces the rate limit using Redis.
func (rl *RedisRateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := ClientIP(r, rl.trustedCIDRs)
		bucket := rl.windowBucket()
		key := fmt.Sprintf("ratelimit:%s:%s", ip, bucket)

		count, err := rl.increment(key)
		if err != nil {
			// Fail open: Redis error → allow the request through.
			logger.Warn("RedisRateLimiter: Redis error — failing open", zap.Error(err), zap.String("key", key))
			next.ServeHTTP(w, r)
			return
		}

		if count > int64(rl.max) {
			retryAfter := int(rl.window.Seconds())
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// increment runs INCR and, if the key is brand new (count == 1), sets EXPIRE.
// It uses a pipeline to minimise round-trips.
func (rl *RedisRateLimiter) increment(key string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pipe := rl.rdb.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	count := incrCmd.Val()
	if count == 1 {
		// Key is new — set expiry. A second error here is acceptable; the key
		// will be cleaned up by Redis eventually.
		expCtx, expCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer expCancel()
		if expErr := rl.rdb.Expire(expCtx, key, rl.window).Err(); expErr != nil {
			logger.Warn("RedisRateLimiter: failed to set EXPIRE", zap.Error(expErr), zap.String("key", key))
		}
	}

	return count, nil
}

// windowBucket returns a string representing the current fixed-window bucket.
// Using Unix time truncated to the window means all requests within the same
// window share the same key.
func (rl *RedisRateLimiter) windowBucket() string {
	now := time.Now().UnixNano()
	windowNs := rl.window.Nanoseconds()
	bucket := now / windowNs
	return strconv.FormatInt(bucket, 10)
}
