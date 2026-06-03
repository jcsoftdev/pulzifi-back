// Package http_test provides smoke tests for the auth module's Chi router.
// Tests confirm that expected routes are registered and reachable (not 404/405)
// without requiring a running database.
//
// CRITICAL: This file tests inline handlers in their current form in module.go.
// Handler extraction is a separate change (extract-inline-handlers).
// TODO(decouple): extract forgot_password and reset_password to own use case dirs.
package http_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	authrepomocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories/mocks"
	authsvcmocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services/mocks"
	authhttp "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/http"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
)

func init() {
	// auth.NewModule() calls config.Load() which fatals on missing required env vars.
	// Set minimal stubs so the module can be constructed in unit tests without a real DB.
	setIfEmpty := func(key, val string) {
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
	setIfEmpty("DB_HOST", "localhost")
	setIfEmpty("DB_PORT", "5432")
	setIfEmpty("DB_NAME", "testdb")
	setIfEmpty("DB_USER", "testuser")
	setIfEmpty("DB_PASSWORD", "testpass")
	setIfEmpty("CORS_ALLOWED_ORIGINS", "http://localhost:3001")
	setIfEmpty("EXTRACTOR_URL", "http://localhost:3005")
}

// ── Stub deps that satisfy interfaces but do nothing ─────────────────────────

type stubNotifier struct{}

func (s *stubNotifier) SendRegistrationSubmitted(_ context.Context, _, _, _ string) error {
	return nil
}

func (s *stubNotifier) SendPasswordReset(_ context.Context, _, _, _ string) error {
	return nil
}

type stubRoleRepo struct{}

func (r *stubRoleRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.Role, error) {
	return nil, nil
}
func (r *stubRoleRepo) GetByName(_ context.Context, _ string) (*entities.Role, error) { return nil, nil }
func (r *stubRoleRepo) GetUserRoles(_ context.Context, _ uuid.UUID) ([]*entities.Role, error) {
	return nil, nil
}
func (r *stubRoleRepo) GetRolePermissions(_ context.Context, _ uuid.UUID) ([]*entities.Permission, error) {
	return nil, nil
}
func (r *stubRoleRepo) AssignRoleToUser(_ context.Context, _, _ uuid.UUID) error { return nil }

type stubPermRepo struct{}

func (p *stubPermRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.Permission, error) {
	return nil, nil
}
func (p *stubPermRepo) GetByName(_ context.Context, _ string) (*entities.Permission, error) {
	return nil, nil
}
func (p *stubPermRepo) GetUserPermissions(_ context.Context, _ uuid.UUID) ([]*entities.Permission, error) {
	return nil, nil
}
func (p *stubPermRepo) HasPermission(_ context.Context, _ uuid.UUID, _, _ string) (bool, error) {
	return false, nil
}

var _ repositories.RoleRepository = (*stubRoleRepo)(nil)
var _ repositories.PermissionRepository = (*stubPermRepo)(nil)
var _ services.RegistrationNotifier = (*stubNotifier)(nil)

// buildAuthRouter constructs a chi.Router with the auth module mounted under a
// parent router that includes Recoverer middleware. Recoverer converts panics
// to HTTP 500 instead of crashing the test process.
//
// This is necessary because some inline handlers in the auth module may nil-deref
// when called with empty/no-op mocks. The smoke tests only verify that routes are
// REGISTERED (not 404), not that they succeed.
func buildAuthRouter(t *testing.T) http.Handler {
	t.Helper()

	userRepo := &authrepomocks.MockUserRepository{}
	rtRepo := &authrepomocks.MockRefreshTokenRepository{}
	tokenSvc := &authsvcmocks.MockTokenService{}
	orgDir := &authsvcmocks.MockOrganizationDirectory{}
	authSvc := &authsvcmocks.MockAuthService{}
	notifier := &stubNotifier{}
	bus := eventbus.GetInstance()

	mod := authhttp.NewModule(authhttp.ModuleDeps{
		UserRepo:         userRepo,
		RefreshTokenRepo: rtRepo,
		RoleRepo:     &stubRoleRepo{},
		PermRepo:     &stubPermRepo{},
		OrgDirectory: orgDir,
		AuthService:      authSvc,
		TokenService:     tokenSvc,
		Notifier:         notifier,
		EventBus:         bus,
	})

	// Wrap in a parent router with Recoverer so panics from inline handlers
	// (nil mock responses) are caught and returned as HTTP 500 rather than
	// crashing the test process.
	parent := chi.NewRouter()
	parent.Use(chiRecoverer)
	mod.RegisterHTTPRoutes(parent)
	return parent
}

// chiRecoverer is a minimal panic recovery middleware that converts panics to HTTP 500.
func chiRecoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// TestAuthRouter_LoginRoute_Reachable verifies POST /auth/login is registered
// and responds (not 404 or 405).
func TestAuthRouter_LoginRoute_Reachable(t *testing.T) {
	handler := buildAuthRouter(t)

	body := bytes.NewBufferString(`{"email":"test@example.com","password":"pw"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code == http.StatusNotFound {
		t.Errorf("POST /auth/login returned 404 — route not registered")
	}
	if rr.Code == http.StatusMethodNotAllowed {
		t.Errorf("POST /auth/login returned 405 — wrong HTTP method registered")
	}
}

// TestAuthRouter_ForgotPasswordRoute_Reachable verifies that the inline handler
// for POST /auth/forgot-password is registered and reachable.
// EXTRACTION JUSTIFIED: forgot_password is currently an inline handler in module.go.
// This test validates the route without requiring handler extraction.
func TestAuthRouter_ForgotPasswordRoute_Reachable(t *testing.T) {
	handler := buildAuthRouter(t)

	body := bytes.NewBufferString(`{"email":"test@example.com"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code == http.StatusNotFound {
		t.Errorf("POST /auth/forgot-password returned 404 — route not registered")
	}
}

// TestAuthRouter_RegisterRoute_Reachable verifies the register endpoint is mounted.
func TestAuthRouter_RegisterRoute_Reachable(t *testing.T) {
	handler := buildAuthRouter(t)

	body := bytes.NewBufferString(`{"email":"new@example.com","password":"pw","firstName":"Joe","lastName":"Doe","subdomain":"acme","organizationName":"Acme Inc"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code == http.StatusNotFound {
		t.Errorf("POST /auth/register returned 404 — route not registered")
	}
}

// TestAuthRouter_LogoutRoute_Reachable verifies the logout endpoint is mounted.
func TestAuthRouter_LogoutRoute_Reachable(t *testing.T) {
	handler := buildAuthRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code == http.StatusNotFound {
		t.Errorf("POST /auth/logout returned 404 — route not registered")
	}
}
