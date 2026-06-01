package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// bodyEchoHandler echoes back the request body so we can verify it was restored.
func bodyEchoHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	})
}

func TestAuthLimiter_BodyRoundTrip(t *testing.T) {
	al := NewAuthLimiter(NewRateLimiter(100, time.Minute), nil)
	wrapped := al.KeyedHandler("login")(bodyEchoHandler())

	body := `{"email":"user@example.com","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "1.2.3.4:1234"
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	got := rec.Body.String()
	if got != body {
		t.Errorf("body not restored: got %q, want %q", got, body)
	}
}

func TestAuthLimiter_BlocksAfterN(t *testing.T) {
	// Allow 2 requests per minute per key.
	al := NewAuthLimiter(NewRateLimiter(2, time.Minute), nil)
	wrapped := al.KeyedHandler("login")(okHandler())

	body := `{"email":"user@example.com"}`

	makeReq := func() int {
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
		return rec.Code
	}

	if got := makeReq(); got != http.StatusOK {
		t.Fatalf("request 1: expected 200, got %d", got)
	}
	if got := makeReq(); got != http.StatusOK {
		t.Fatalf("request 2: expected 200, got %d", got)
	}
	if got := makeReq(); got != http.StatusTooManyRequests {
		t.Fatalf("request 3: expected 429, got %d", got)
	}
}

func TestAuthLimiter_DifferentEmailsIsolated(t *testing.T) {
	al := NewAuthLimiter(NewRateLimiter(1, time.Minute), nil)
	wrapped := al.KeyedHandler("login")(okHandler())

	sendAs := func(email string) int {
		body := `{"email":"` + email + `"}`
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
		req.RemoteAddr = "1.2.3.4:1234"
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
		return rec.Code
	}

	// Exhaust userA.
	if got := sendAs("usera@example.com"); got != http.StatusOK {
		t.Fatalf("userA req 1: expected 200, got %d", got)
	}
	if got := sendAs("usera@example.com"); got != http.StatusTooManyRequests {
		t.Fatalf("userA req 2: expected 429, got %d", got)
	}

	// userB should not be affected.
	if got := sendAs("userb@example.com"); got != http.StatusOK {
		t.Fatalf("userB: expected 200, got %d", got)
	}
}

func TestAuthLimiter_BlankEmailFallsBackToIP(t *testing.T) {
	// With blank email the limiter falls back to IP-only key.
	al := NewAuthLimiter(NewRateLimiter(1, time.Minute), nil)
	wrapped := al.KeyedHandler("login")(okHandler())

	sendBlank := func() int {
		body := `{"email":""}`
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
		req.RemoteAddr = "5.5.5.5:1234"
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
		return rec.Code
	}

	if got := sendBlank(); got != http.StatusOK {
		t.Fatalf("blank email req 1: expected 200, got %d", got)
	}
	if got := sendBlank(); got != http.StatusTooManyRequests {
		t.Fatalf("blank email req 2: expected 429, got %d", got)
	}
}

func TestAuthLimiter_NoBodyFallsBackToIP(t *testing.T) {
	al := NewAuthLimiter(NewRateLimiter(1, time.Minute), nil)
	wrapped := al.KeyedHandler("login")(okHandler())

	sendNoBody := func() int {
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "7.7.7.7:1234"
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)
		return rec.Code
	}

	if got := sendNoBody(); got != http.StatusOK {
		t.Fatalf("no body req 1: expected 200, got %d", got)
	}
	if got := sendNoBody(); got != http.StatusTooManyRequests {
		t.Fatalf("no body req 2: expected 429, got %d", got)
	}
}
