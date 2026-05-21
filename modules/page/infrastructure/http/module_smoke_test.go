package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	pagehttp "github.com/jcsoftdev/pulzifi-back/modules/page/infrastructure/http"
)

// TestPageModuleSmoke verifies that GET /pages is registered (not 404).
// NewModule() creates the module without a DB; the route must exist even if the
// handler returns a non-200 response due to missing dependencies.
func TestPageModuleSmoke(t *testing.T) {
	mod := pagehttp.NewModule() // no DB

	r := chi.NewRouter()
	mod.RegisterHTTPRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/pages", http.NoBody)
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
		t.Errorf("GET /pages returned 404 — route is not registered")
	}
}
