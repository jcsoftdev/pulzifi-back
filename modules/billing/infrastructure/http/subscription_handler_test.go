// Package http (internal tests) verifies HTTP handler internals using whitebox access.
// This file is in the same package (not _test) so it can call private methods.
package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	getsubscription "github.com/jcsoftdev/pulzifi-back/modules/billing/application/get_subscription"
	"github.com/jcsoftdev/pulzifi-back/shared/contextkeys"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/persistence/inmem"
)

// TestHandleGetSubscription_NoSubscriptionReturns200 verifies FR7:
// when an org has no Stripe subscription the endpoint must return HTTP 200
// with stripe_status: null — NOT 404.
//
// RED: currently returns 404 when ErrSubscriptionNotFound.
// GREEN: will return 200 with null stripe fields.
func TestHandleGetSubscription_NoSubscriptionReturns200(t *testing.T) {
	// Empty subscription repo — no Stripe subscription exists for any org.
	subRepo := inmem.NewSubscriptionRepo()
	subHandler := getsubscription.NewHandler(subRepo)

	mod := &Module{
		deps: Deps{
			SubscriptionHandler: subHandler,
		},
	}

	// Call handleGetSubscription directly — bypass auth/DB middleware.
	// We inject an empty org ID to trigger ErrSubscriptionNotFound.
	// The handler calls SubscriptionHandler.Handle(ctx, orgID) with an empty string
	// which causes ErrSubscriptionNotFound (invalid UUID path).
	req := httptest.NewRequest(http.MethodGet, "/billing/subscription", http.NoBody)
	rr := httptest.NewRecorder()

	// Provide a minimal context: no org context, which will produce ErrSubscriptionNotFound
	// via getsubscription.Handle("") — the use case treats bad UUID as not found.
	mod.handleGetSubscriptionWithOrgID(rr, req, "")

	if rr.Code != http.StatusOK {
		t.Fatalf("FR7: expected 200 for no-subscription org, got %d (body: %s)", rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("could not decode response: %v", err)
	}

	// stripe_status must be null (JSON null = Go nil in map[string]interface{}).
	if v, ok := resp["stripe_status"]; !ok || v != nil {
		t.Errorf("expected stripe_status: null, got key_present=%v value=%v", ok, v)
	}
}

// ── NFR3 Role Enforcement Tests ───────────────────────────────────────────────

// roleContext injects roles into a request context to simulate the output of
// AuthMiddleware.Authenticate (which sets UserRolesKey from JWT claims).
func roleContext(req *http.Request, roles []string) *http.Request {
	ctx := context.WithValue(req.Context(), contextkeys.UserRolesKey, roles)
	return req.WithContext(ctx)
}

// TestCheckoutPortal_RequiresAdminRole verifies NFR3:
// POST /billing/checkout and POST /billing/portal MUST return 403 when the
// caller's JWT contains only USER role (a regular org member).
func TestCheckoutPortal_RequiresAdminRole(t *testing.T) {
	mod := &Module{deps: Deps{}}

	requireAdmin := mod.requireAdminOrSuperAdmin()

	tests := []struct {
		name      string
		roles     []string
		wantCode  int
	}{
		{
			name:     "USER role gets 403",
			roles:    []string{"USER"},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "VIEWER role gets 403",
			roles:    []string{"VIEWER"},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "no roles gets 403",
			roles:    []string{},
			wantCode: http.StatusForbidden,
		},
		{
			name:     "ADMIN role gets 200",
			roles:    []string{"ADMIN"},
			wantCode: http.StatusOK,
		},
		{
			name:     "SUPER_ADMIN role gets 200",
			roles:    []string{"SUPER_ADMIN"},
			wantCode: http.StatusOK,
		},
		{
			name:     "USER + ADMIN gets 200 (role escalation)",
			roles:    []string{"USER", "ADMIN"},
			wantCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Build a minimal router with our admin-only middleware.
			r := chi.NewRouter()
			r.With(requireAdmin).Post("/billing/checkout", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodPost, "/billing/checkout", http.NoBody)
			req = roleContext(req, tc.roles)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tc.wantCode {
				t.Errorf("roles=%v: expected %d, got %d (body: %s)", tc.roles, tc.wantCode, rr.Code, rr.Body.String())
			}
		})
	}
}
