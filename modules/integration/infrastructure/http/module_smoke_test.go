package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	integrationhttp "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/http"
)

// TestIntegrationModuleSmoke verifies that GET /integrations is registered (not 404).
// NewModuleWithDB(nil) creates the module without dependencies; auth middleware is nil
// so it may panic or return non-200 — but neither is 404.
func TestIntegrationModuleSmoke(t *testing.T) {
	mod := integrationhttp.NewModuleWithDB(nil)

	r := chi.NewRouter()
	mod.RegisterHTTPRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/integrations", http.NoBody)
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
		t.Errorf("GET /integrations returned 404 — route is not registered")
	}
}
