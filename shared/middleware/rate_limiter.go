package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type visitor struct {
	tokens   int
	lastSeen time.Time
	mu       sync.Mutex
}

// RateLimiter implements an in-memory token bucket rate limiter per IP.
type RateLimiter struct {
	visitors      sync.Map
	maxTokens     int
	window        time.Duration
	quit          chan struct{}
	trustedCIDRs  []net.IPNet
}

// NewRateLimiter creates a rate limiter that allows maxTokens requests per window per IP.
// IP extraction uses RemoteAddr directly (no trusted-proxy header processing).
// Use NewRateLimiterWithTrustedProxies when running behind a proxy.
func NewRateLimiter(maxTokens int, window time.Duration) *RateLimiter {
	return NewRateLimiterWithTrustedProxies(maxTokens, window, nil)
}

// NewRateLimiterWithTrustedProxies creates a rate limiter that respects the given
// trusted CIDR list when extracting the real client IP from X-Forwarded-For.
// Pass nil or an empty slice to fall back to RemoteAddr-only behaviour (same as NewRateLimiter).
func NewRateLimiterWithTrustedProxies(maxTokens int, window time.Duration, trusted []net.IPNet) *RateLimiter {
	rl := &RateLimiter{
		maxTokens:    maxTokens,
		window:       window,
		quit:         make(chan struct{}),
		trustedCIDRs: trusted,
	}
	go rl.cleanup()
	return rl
}

// Stop terminates the background cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.quit)
}

// Handler returns an http.Handler middleware that enforces the rate limit.
func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := ClientIP(r, rl.trustedCIDRs)

		v := rl.getVisitor(ip)
		v.mu.Lock()

		now := time.Now()
		elapsed := now.Sub(v.lastSeen)

		// Replenish tokens based on elapsed time
		if elapsed >= rl.window {
			v.tokens = rl.maxTokens
		} else {
			replenish := int(float64(rl.maxTokens) * (float64(elapsed) / float64(rl.window)))
			v.tokens += replenish
			if v.tokens > rl.maxTokens {
				v.tokens = rl.maxTokens
			}
		}
		v.lastSeen = now

		if v.tokens <= 0 {
			v.mu.Unlock()
			retryAfter := int(rl.window.Seconds())
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		v.tokens--
		v.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) getVisitor(ip string) *visitor {
	if v, ok := rl.visitors.Load(ip); ok {
		return v.(*visitor)
	}
	v := &visitor{tokens: rl.maxTokens, lastSeen: time.Now()}
	actual, _ := rl.visitors.LoadOrStore(ip, v)
	return actual.(*visitor)
}

// cleanup removes stale visitor entries every window interval.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			threshold := time.Now().Add(-2 * rl.window)
			rl.visitors.Range(func(key, value interface{}) bool {
				v := value.(*visitor)
				v.mu.Lock()
				stale := v.lastSeen.Before(threshold)
				v.mu.Unlock()
				if stale {
					rl.visitors.Delete(key)
				}
				return true
			})
		case <-rl.quit:
			return
		}
	}
}
