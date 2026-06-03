// Package http_test provides smoke tests for the POST /auth/onboarding endpoint.
// These tests verify that:
//   - The route is registered and requires authentication (401 without token)
//   - Invalid body yields 400
//   - Happy path returns 200 with the correct response shape
//   - Already-provisioned user receives 409
package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	autherrors "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/errors"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	authrepomocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories/mocks"
	authsvcmocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services/mocks"
	authhttp "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/http"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
)

// buildOnboardingRouter returns a Chi router WITHOUT a valid token service
// configuration (ValidateToken returns nil) so unauthenticated requests are
// naturally rejected 401 by the auth middleware.
func buildOnboardingRouter(t *testing.T) http.Handler {
	t.Helper()

	bus := eventbus.GetInstance()
	membershipChecker := &authsvcmocks.MockOrganizationMembershipChecker{}
	trialProvisioner := &authsvcmocks.MockTrialProvisioner{}

	// ValidateTokenResult=nil means any token attempt will get a nil claim → 401.
	tokenSvc := &authsvcmocks.MockTokenService{ValidateTokenResult: nil, ValidateTokenErr: errFakeUnauthorized}

	mod := authhttp.NewModule(authhttp.ModuleDeps{
		UserRepo:          &authrepomocks.MockUserRepository{},
		RefreshTokenRepo:  &authrepomocks.MockRefreshTokenRepository{},
		RoleRepo:          &stubRoleRepo{},
		PermRepo:          &stubPermRepo{},
		OrgDirectory:      &authsvcmocks.MockOrganizationDirectory{},
		AuthService:       &authsvcmocks.MockAuthService{},
		TokenService:      tokenSvc,
		MembershipChecker: membershipChecker,
		TrialProvisioner:  trialProvisioner,
		Notifier:          &stubNotifier{},
		EventBus:          bus,
	})

	parent := chi.NewRouter()
	parent.Use(chiRecoverer)
	mod.RegisterHTTPRoutes(parent)
	return parent
}

// errFakeUnauthorized is returned by MockTokenService.ValidateToken when we
// want the auth middleware to reject the request.
var errFakeUnauthorized = fmt.Errorf("invalid token")

// buildOnboardingRouterWithToken returns a Chi router whose MockTokenService
// returns valid claims for any token string. Callers must include an
// "Authorization: Bearer fake-token" header to satisfy the auth middleware.
func buildOnboardingRouterWithToken(
	t *testing.T,
	userID uuid.UUID,
	membershipResult bool,
	membershipErr error,
) http.Handler {
	t.Helper()

	bus := eventbus.GetInstance()

	// Configure MockTokenService to return valid claims for any token string.
	tokenSvc := &authsvcmocks.MockTokenService{
		ValidateTokenResult: &authservices.TokenClaims{
			UserID: userID,
			Email:  "test@example.com",
		},
	}

	membershipChecker := &authsvcmocks.MockOrganizationMembershipChecker{
		HasAnyMembershipResult: membershipResult,
		HasAnyMembershipErr:    membershipErr,
	}
	trialProvisioner := &authsvcmocks.MockTrialProvisioner{
		ProvisionResult: &authservices.TrialProvisionOutput{
			OrganizationID:        uuid.New(),
			OrganizationSubdomain: "acme",
			SchemaName:            "tenant_acme",
			TrialEndsAt:           time.Now().Add(14 * 24 * time.Hour),
		},
	}

	mod := authhttp.NewModule(authhttp.ModuleDeps{
		UserRepo:          &authrepomocks.MockUserRepository{},
		RefreshTokenRepo:  &authrepomocks.MockRefreshTokenRepository{},
		RoleRepo:          &stubRoleRepo{},
		PermRepo:          &stubPermRepo{},
		OrgDirectory:      &authsvcmocks.MockOrganizationDirectory{},
		AuthService:       &authsvcmocks.MockAuthService{},
		TokenService:      tokenSvc,
		MembershipChecker: membershipChecker,
		TrialProvisioner:  trialProvisioner,
		Notifier:          &stubNotifier{},
		EventBus:          bus,
	})

	parent := chi.NewRouter()
	parent.Use(chiRecoverer)
	mod.RegisterHTTPRoutes(parent)
	return parent
}

// addAuth sets the Authorization header on a request so the auth middleware
// picks up a Bearer token (validated by MockTokenService above).
func addAuth(req *http.Request) *http.Request {
	req.Header.Set("Authorization", "Bearer fake-token")
	return req
}

// TestOnboarding_401_WithoutAuth verifies that POST /auth/onboarding returns
// 401 when no authentication token is present.
func TestOnboarding_401_WithoutAuth(t *testing.T) {
	handler := buildOnboardingRouter(t)

	body := bytes.NewBufferString(`{"org_name":"Acme","subdomain":"acme"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/onboarding", body)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d", rr.Code)
	}
}

// TestOnboarding_400_EmptyBody verifies that POST /auth/onboarding returns 400
// when the request body is missing required fields.
func TestOnboarding_400_EmptyBody(t *testing.T) {
	userID := uuid.New()
	handler := buildOnboardingRouterWithToken(t, userID, false, nil)

	body := bytes.NewBufferString(`{"org_name":"","subdomain":""}`)
	req := addAuth(httptest.NewRequest(http.MethodPost, "/auth/onboarding", body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", rr.Code)
	}
}

// TestOnboarding_400_InvalidSubdomain verifies that a subdomain with invalid
// characters is rejected with 400.
func TestOnboarding_400_InvalidSubdomain(t *testing.T) {
	userID := uuid.New()
	handler := buildOnboardingRouterWithToken(t, userID, false, nil)

	body := bytes.NewBufferString(`{"org_name":"Acme Inc","subdomain":"my org!"}`)
	req := addAuth(httptest.NewRequest(http.MethodPost, "/auth/onboarding", body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid subdomain, got %d", rr.Code)
	}
}

// TestOnboarding_200_HappyPath verifies that POST /auth/onboarding returns 200
// with a correct response envelope when the user has no existing membership.
func TestOnboarding_200_HappyPath(t *testing.T) {
	userID := uuid.New()
	// membershipResult=false → user has no org → provisioner will be called
	handler := buildOnboardingRouterWithToken(t, userID, false, nil)

	body := bytes.NewBufferString(`{"org_name":"Acme Inc","subdomain":"acme"}`)
	req := addAuth(httptest.NewRequest(http.MethodPost, "/auth/onboarding", body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", rr.Code, rr.Body.String())
		return
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := resp["subdomain"]; !ok {
		t.Errorf("response missing 'subdomain' field: %v", resp)
	}
	if _, ok := resp["trial_ends_at"]; !ok {
		t.Errorf("response missing 'trial_ends_at' field: %v", resp)
	}
}

// TestOnboarding_409_AlreadyProvisioned verifies that POST /auth/onboarding
// returns 409 when the membership checker reports the user already has an org.
func TestOnboarding_409_AlreadyProvisioned(t *testing.T) {
	userID := uuid.New()
	// membershipResult=true → ErrAlreadyProvisioned path
	handler := buildOnboardingRouterWithToken(t, userID, true, nil)

	body := bytes.NewBufferString(`{"org_name":"Acme Inc","subdomain":"acme"}`)
	req := addAuth(httptest.NewRequest(http.MethodPost, "/auth/onboarding", body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body: %s)", rr.Code, rr.Body.String())
		return
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["error"] != "already_provisioned" {
		t.Errorf("expected error=already_provisioned, got %q", resp["error"])
	}
}

// Compile-time check: ErrAlreadyProvisioned is used by the handler error-mapping.
var _ = autherrors.ErrAlreadyProvisioned
