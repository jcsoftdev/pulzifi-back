package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	emailhttp "github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/http"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
)

// noopEmailProvider is a no-op implementation of services.EmailProvider for tests.
type noopEmailProvider struct{}

func (noopEmailProvider) Send(_ context.Context, _, _, _ string) error { return nil }

var _ emailservices.EmailProvider = noopEmailProvider{}

// TestEmailModuleSmoke verifies that POST /emails/send is registered (not 404).
func TestEmailModuleSmoke(t *testing.T) {
	mod := emailhttp.NewModule(noopEmailProvider{})

	r := chi.NewRouter()
	mod.RegisterHTTPRoutes(r)

	body := `{"to":"test@example.com","subject":"hi","body":"hello"}`
	req := httptest.NewRequest(http.MethodPost, "/emails/send", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	panicked := false
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				panicked = true
			}
		}()
		r.ServeHTTP(rr, req)
	}()

	if !panicked && rr.Code == http.StatusNotFound {
		t.Errorf("POST /emails/send returned 404 — route is not registered")
	}
}
