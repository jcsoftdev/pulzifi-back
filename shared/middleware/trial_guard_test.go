package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeTrialReader struct {
	expired bool
	err     error
	calls   int
}

func (f *fakeTrialReader) TrialExpired(_ context.Context, _ string) (bool, error) {
	f.calls++
	return f.expired, f.err
}

func trialGuardOKHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

// reqWithSubdomain returns a request whose context has the tenant subdomain set.
func reqWithSubdomain(method, target, subdomain string) *http.Request {
	r := httptest.NewRequest(method, target, nil)
	ctx := context.WithValue(r.Context(), SubdomainContextKey, subdomain)
	return r.WithContext(ctx)
}

func TestTrialGuard_PassesReadMethods(t *testing.T) {
	reader := &fakeTrialReader{expired: true}
	guard := NewTrialGuard(reader)
	handler := guard.Handler(trialGuardOKHandler())

	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		t.Run(method, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, reqWithSubdomain(method, "/api/v1/pages", "acme"))
			if rr.Code != http.StatusOK {
				t.Fatalf("read-method %s: expected 200, got %d", method, rr.Code)
			}
		})
	}
}

func TestTrialGuard_BlocksWritesWhenExpired(t *testing.T) {
	reader := &fakeTrialReader{expired: true}
	guard := NewTrialGuard(reader)
	handler := guard.Handler(trialGuardOKHandler())

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, reqWithSubdomain(method, "/api/v1/pages", "acme"))
			if rr.Code != http.StatusPaymentRequired {
				t.Fatalf("write-method %s: expected 402, got %d", method, rr.Code)
			}
		})
	}
}

func TestTrialGuard_AllowsCheckoutEvenWhenExpired(t *testing.T) {
	reader := &fakeTrialReader{expired: true}
	guard := NewTrialGuard(reader)
	handler := guard.Handler(trialGuardOKHandler())

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, reqWithSubdomain(http.MethodPost, "/api/v1/billing/checkout", "acme"))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 on /billing/checkout despite expired trial, got %d", rr.Code)
	}
}

func TestTrialGuard_PassesWritesWhenNotExpired(t *testing.T) {
	reader := &fakeTrialReader{expired: false}
	guard := NewTrialGuard(reader)
	handler := guard.Handler(trialGuardOKHandler())

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, reqWithSubdomain(http.MethodPost, "/api/v1/pages", "acme"))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 when not expired, got %d", rr.Code)
	}
}

func TestTrialGuard_FailsOpenOnReaderError(t *testing.T) {
	reader := &fakeTrialReader{err: errors.New("db lost")}
	guard := NewTrialGuard(reader)
	handler := guard.Handler(trialGuardOKHandler())

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, reqWithSubdomain(http.MethodPost, "/api/v1/pages", "acme"))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 (fail-open) on reader error, got %d", rr.Code)
	}
}

func TestTrialGuard_NoSubdomainPassesThrough(t *testing.T) {
	reader := &fakeTrialReader{expired: true}
	guard := NewTrialGuard(reader)
	handler := guard.Handler(trialGuardOKHandler())

	rr := httptest.NewRecorder()
	// No tenant in context
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected pass-through without tenant, got %d", rr.Code)
	}
	if reader.calls != 0 {
		t.Errorf("expected reader not to be called without tenant, got %d calls", reader.calls)
	}
}
