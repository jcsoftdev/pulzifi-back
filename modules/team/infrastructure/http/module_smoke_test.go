package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	teamhttp "github.com/jcsoftdev/pulzifi-back/modules/team/infrastructure/http"
	teamservices "github.com/jcsoftdev/pulzifi-back/modules/team/domain/services"
)

// ── no-op stubs ───────────────────────────────────────────────────────────────

type noopTeamEmailer struct{}

func (noopTeamEmailer) Send(_ context.Context, _, _, _ string) error { return nil }

var _ teamservices.Emailer = noopTeamEmailer{}

type noopInviteEmailBuilder struct{}

func (noopInviteEmailBuilder) TeamInvite(_, _, _ string) (string, string) { return "subj", "<html/>" }

var _ teamservices.InviteEmailBuilder = noopInviteEmailBuilder{}

type noopTokenGenerator struct{}

func (noopTokenGenerator) Generate(_ context.Context, _ uuid.UUID) (string, error) {
	return "test-token", nil
}

var _ teamservices.TokenGenerator = noopTokenGenerator{}

// TestTeamModuleSmoke verifies that GET /team/members is registered (not 404).
// NewModuleWithDB(nil, ...) creates the module without a real DB connection;
// the route must exist even if the handler returns non-200 due to nil DB.
func TestTeamModuleSmoke(t *testing.T) {
	mod := teamhttp.NewModuleWithDB(
		nil, // no real DB needed for route registration check
		noopTeamEmailer{},
		noopInviteEmailBuilder{},
		noopTokenGenerator{},
		"http://localhost:3001",
	)

	r := chi.NewRouter()
	mod.RegisterHTTPRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/team/members", http.NoBody)
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
		t.Errorf("GET /team/members returned 404 — route is not registered")
	}
}
