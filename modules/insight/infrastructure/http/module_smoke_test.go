package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	insighthttp "github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/http"
)

// TestInsightModuleSmoke verifies that GET /insights is registered (not 404).
// NewModule() creates the module without a DB; the handler returns 500 or panics
// on nil auth middleware, but either is preferable to 404.
func TestInsightModuleSmoke(t *testing.T) {
	mod := insighthttp.NewModule() // no DB

	r := chi.NewRouter()
	mod.RegisterHTTPRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/insights", http.NoBody)
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
		t.Errorf("GET /insights returned 404 — route is not registered")
	}
}
