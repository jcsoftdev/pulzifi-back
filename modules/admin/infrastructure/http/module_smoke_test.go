package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	adminhttp "github.com/jcsoftdev/pulzifi-back/modules/admin/infrastructure/http"
	adminservices "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories/mocks"
	sharedmw "github.com/jcsoftdev/pulzifi-back/shared/middleware"
)

// ── no-op stubs ───────────────────────────────────────────────────────────────

type noopApprovalProvisioner struct{}

func (noopApprovalProvisioner) Provision(_ context.Context, _ adminservices.ApprovalInput) error {
	return nil
}

type noopRejectionProvisioner struct{}

func (noopRejectionProvisioner) Reject(_ context.Context, _ adminservices.RejectionInput) error {
	return nil
}

type noopPendingUserReader struct{}

func (noopPendingUserReader) GetByID(_ context.Context, _ uuid.UUID) (*adminservices.PendingUser, error) {
	return nil, nil
}

type noopRegistrationNotifier struct{}

func (noopRegistrationNotifier) SendApproval(_ context.Context, _, _, _, _ string) error { return nil }
func (noopRegistrationNotifier) SendRejection(_ context.Context, _, _ string) error      { return nil }

type noopAuthVerifier struct{}

func (noopAuthVerifier) Authenticate(next http.Handler) http.Handler { return next }
func (noopAuthVerifier) RequirePermission(_, _ string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { return next }
}
func (noopAuthVerifier) RequireRole(_ string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler { return next }
}

// Verify the stub implements the interface at compile-time.
var _ sharedmw.AuthVerifier = noopAuthVerifier{}

// TestAdminModuleSmoke verifies that GET /admin/users/pending is registered (not 404).
func TestAdminModuleSmoke(t *testing.T) {
	deps := adminhttp.ModuleDeps{
		RegReqRepo:           &mocks.MockRegistrationRequestRepository{},
		UserReader:           noopPendingUserReader{},
		ApprovalProvisioner:  noopApprovalProvisioner{},
		RejectionProvisioner: noopRejectionProvisioner{},
		Notifier:             noopRegistrationNotifier{},
		AuthMiddleware:       noopAuthVerifier{},
		FrontendURL:          "http://localhost:3001",
	}
	mod := adminhttp.NewModule(deps)

	r := chi.NewRouter()
	mod.RegisterHTTPRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/admin/users/pending", http.NoBody)
	rr := httptest.NewRecorder()

	func() {
		defer func() { recover() }() //nolint:errcheck
		r.ServeHTTP(rr, req)
	}()

	if rr.Code == http.StatusNotFound {
		t.Errorf("GET /admin/users/pending returned 404 — route is not registered")
	}
}
