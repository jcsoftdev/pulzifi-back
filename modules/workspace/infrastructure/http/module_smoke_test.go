package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	workspacehttp "github.com/jcsoftdev/pulzifi-back/modules/workspace/infrastructure/http"
)

// TestWorkspaceModuleSmoke verifies that GET /workspaces is registered (not 404).
// NewModule() creates the module without a DB; the handler returns 200 with an
// empty list (per the nil-DB fast-path in module.go), confirming the route exists.
func TestWorkspaceModuleSmoke(t *testing.T) {
	mod := workspacehttp.NewModule() // nil DB → returns mock 200 for GET /workspaces

	r := chi.NewRouter()
	mod.RegisterHTTPRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/workspaces", http.NoBody)
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
		t.Errorf("GET /workspaces returned 404 — route is not registered")
	}
}
