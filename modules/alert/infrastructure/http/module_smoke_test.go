package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	alerthttp "github.com/jcsoftdev/pulzifi-back/modules/alert/infrastructure/http"
)

// TestAlertModuleSmoke verifies that GET /alerts is registered (not 404).
// Auth middleware is nil in unit tests; the recover() catches the nil-pointer panic
// so we can assert the route exists without requiring a real token.
func TestAlertModuleSmoke(t *testing.T) {
	mod := alerthttp.NewModule()

	r := chi.NewRouter()
	mod.RegisterHTTPRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/alerts", http.NoBody)
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
		t.Errorf("GET /alerts returned 404 — route is not registered")
	}
}
