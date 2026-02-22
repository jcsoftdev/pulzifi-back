package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type visitor struct {
	tokens    int
	lastSeen  time.Time
	mu        sync.Mutex
}

// RateLimiter implements an in-memory token bucket rate limiter per IP.
type RateLimiter struct {
	visitors   sync.Map
	maxTokens  int
	window     time.Duration
	quit       chan struct{}
}

// NewRateLimiter creates a rate limiter that allows maxTokens requests per window per IP.
func NewRateLimiter(maxTokens int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		maxTokens: maxTokens,
		window:    window,
		quit:      make(chan struct{}),
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
		ip := extractIP(r)

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

// extractIP returns the client IP from the request.
func extractIP(r *http.Request) string {
	// Check X-Forwarded-For first (may contain comma-separated list)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
