package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	dashhttp "github.com/jcsoftdev/pulzifi-back/modules/dashboard/infrastructure/http"
)

// TestDashboardModuleSmoke verifies that GET /dashboard/stats is registered (not 404).
// NewModule() creates the module without a DB, so the handler returns 500 (db not initialized)
// rather than 404 — confirming the route IS registered.
func TestDashboardModuleSmoke(t *testing.T) {
	mod := dashhttp.NewModule() // no DB — returns 500 on handler, not 404

	r := chi.NewRouter()
	mod.RegisterHTTPRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/stats", http.NoBody)
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
		t.Errorf("GET /dashboard/stats returned 404 — route is not registered")
	}
}
