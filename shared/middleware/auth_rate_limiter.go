package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

// AuthLimiter is a per-route rate limiter keyed by (lowercased email, clientIP).
// It wraps any Limiter (memory or Redis) and intercepts the request body to
// extract the email field, restoring the body for downstream handlers.
//
// Key format: authrl:<routeTag>:<lowercased-email>:<clientIP>
// When email is blank or the body cannot be decoded, the key falls back to:
//   authrl:<routeTag>:<clientIP>
//
// The limiter fails open on backend errors (inherited from the underlying Limiter).
type AuthLimiter struct {
	inner        Limiter
	trustedCIDRs []net.IPNet
}

// NewAuthLimiter creates an AuthLimiter wrapping the provided inner Limiter.
// trusted is passed to ClientIP for real-IP extraction.
func NewAuthLimiter(inner Limiter, trusted []net.IPNet) *AuthLimiter {
	return &AuthLimiter{
		inner:        inner,
		trustedCIDRs: trusted,
	}
}

// KeyedHandler returns a middleware that rate-limits based on routeTag + email + clientIP.
// Typical usage:
//
//	router.With(al.KeyedHandler("login")).Post("/login", loginHandler)
func (al *AuthLimiter) KeyedHandler(routeTag string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := ClientIP(r, al.trustedCIDRs)
			email := peekEmail(r) // also restores r.Body

			key := buildAuthKey(routeTag, email, clientIP)

			// Wrap the next handler behind a synthetic request carrying the key
			// as a synthetic X-Forwarded-For so the inner Limiter keys on it.
			// This avoids coupling to a specific Limiter implementation.
			// Instead, we use a keyed in-process approach: clone the request,
			// set RemoteAddr to the compound key, delegate to inner.Handler.
			//
			// For the memory limiter the key becomes the "IP" in its sync.Map —
			// effectively giving us per-(route+email+ip) buckets.
			// For the Redis limiter the key is used as the Redis key prefix.
			keyedReq := r.Clone(r.Context())
			keyedReq.RemoteAddr = key + ":0" // port required by SplitHostPort

			al.inner.Handler(next).ServeHTTP(w, keyedReq)
		})
	}
}

// peekEmail reads r.Body, decodes the "email" field, and restores r.Body so
// the downstream handler can read it again.  Returns empty string on any error.
func peekEmail(r *http.Request) string {
	if r.Body == nil {
		return ""
	}

	buf, err := io.ReadAll(r.Body)
	// Always restore the body, even on error.
	r.Body = io.NopCloser(bytes.NewReader(buf))
	if err != nil {
		return ""
	}

	var payload struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(buf, &payload); err != nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(payload.Email))
}

// buildAuthKey returns the composite limiter key.
// When email is empty the key omits the email segment (IP-only fallback).
func buildAuthKey(routeTag, email, clientIP string) string {
	if email == "" {
		return fmt.Sprintf("authrl:%s:%s", routeTag, clientIP)
	}
	return fmt.Sprintf("authrl:%s:%s:%s", routeTag, email, clientIP)
}
