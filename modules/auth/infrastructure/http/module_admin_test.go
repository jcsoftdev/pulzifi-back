package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	authrepomocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories/mocks"
	authsvcmocks "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services/mocks"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	authhttp "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/http"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
)

// buildAdminRouterWithTokenSvc constructs a router where the token service
// accepts any Bearer token and returns the provided claims. This lets tests
// satisfy the Authenticate middleware without a real JWT.
func buildAdminRouterWithTokenSvc(
	t *testing.T,
	userRepo *authrepomocks.MockUserRepository,
	roleRepo *stubRoleRepoFull,
	claims *services.TokenClaims,
) http.Handler {
	t.Helper()

	rtRepo := &authrepomocks.MockRefreshTokenRepository{}
	tokenSvc := &authsvcmocks.MockTokenService{
		ValidateTokenResult: claims,
	}
	orgDir := &authsvcmocks.MockOrganizationDirectory{}
	authSvc := &authsvcmocks.MockAuthService{}
	notifier := &stubNotifier{}
	bus := eventbus.GetInstance()

	mod := authhttp.NewModule(authhttp.ModuleDeps{
		UserRepo:         userRepo,
		RefreshTokenRepo: rtRepo,
		RoleRepo:         roleRepo,
		PermRepo:         &stubPermRepo{},
		OrgDirectory:     orgDir,
		AuthService:      authSvc,
		TokenService:     tokenSvc,
		Notifier:         notifier,
		EventBus:         bus,
		// DB is nil → list-users handler not wired (returns 503)
	})

	parent := chi.NewRouter()
	parent.Use(chiRecoverer)
	mod.RegisterHTTPRoutes(parent)
	return parent
}

// superAdminClaims returns TokenClaims for a SUPER_ADMIN caller.
func superAdminClaims() *services.TokenClaims {
	return &services.TokenClaims{
		UserID: uuid.New(),
		Email:  "admin@example.com",
		Roles:  []string{"SUPER_ADMIN"},
	}
}

// nonAdminClaims returns TokenClaims for a non-SUPER_ADMIN caller.
func nonAdminClaims() *services.TokenClaims {
	return &services.TokenClaims{
		UserID: uuid.New(),
		Email:  "user@example.com",
		Roles:  []string{"ADMIN"},
	}
}

// withBearer adds an Authorization header so the Authenticate middleware
// can find a token to pass to ValidateToken.
func withBearer(r *http.Request) *http.Request {
	r.Header.Set("Authorization", "Bearer fake-token")
	return r
}

// stubRoleRepoFull is a configurable in-memory RoleRepository.
type stubRoleRepoFull struct {
	rolesByName   map[string]*entities.Role
	userRoles     map[uuid.UUID][]*entities.Role
	assignedCalls []struct{ userID, roleID uuid.UUID }

	getByIDResult *entities.Role
	assignErr     error
}

func (s *stubRoleRepoFull) GetByID(_ context.Context, _ uuid.UUID) (*entities.Role, error) {
	return s.getByIDResult, nil
}
func (s *stubRoleRepoFull) GetByName(_ context.Context, name string) (*entities.Role, error) {
	if r, ok := s.rolesByName[name]; ok {
		return r, nil
	}
	return nil, nil
}
func (s *stubRoleRepoFull) GetUserRoles(_ context.Context, userID uuid.UUID) ([]*entities.Role, error) {
	return s.userRoles[userID], nil
}
func (s *stubRoleRepoFull) GetRolePermissions(_ context.Context, _ uuid.UUID) ([]*entities.Permission, error) {
	return nil, nil
}
func (s *stubRoleRepoFull) AssignRoleToUser(_ context.Context, userID, roleID uuid.UUID) error {
	if s.assignErr != nil {
		return s.assignErr
	}
	s.assignedCalls = append(s.assignedCalls, struct{ userID, roleID uuid.UUID }{userID, roleID})
	return nil
}

// ── requireSuperAdmin gate ────────────────────────────────────────────────────

func TestAdminRoutes_ForbiddenWithoutSuperAdmin(t *testing.T) {
	userRepo := &authrepomocks.MockUserRepository{}
	roleRepo := &stubRoleRepoFull{}
	handler := buildAdminRouterWithTokenSvc(t, userRepo, roleRepo, nonAdminClaims())

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/auth/admin/users"},
		{http.MethodPost, "/auth/admin/users/" + uuid.New().String() + "/promote"},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req = withBearer(req)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Errorf("expected 403 Forbidden, got %d", rr.Code)
			}
		})
	}
}

// ── Promote endpoint ──────────────────────────────────────────────────────────

func TestAdminPromote_UserNotFound_Returns404(t *testing.T) {
	targetID := uuid.New()
	userRepo := &authrepomocks.MockUserRepository{
		GetByIDUser: nil, // user not found
	}
	superAdminRole := &entities.Role{ID: uuid.New(), Name: "SUPER_ADMIN"}
	roleRepo := &stubRoleRepoFull{
		rolesByName: map[string]*entities.Role{"SUPER_ADMIN": superAdminRole},
		userRoles:   map[uuid.UUID][]*entities.Role{},
	}
	handler := buildAdminRouterWithTokenSvc(t, userRepo, roleRepo, superAdminClaims())

	req := httptest.NewRequest(http.MethodPost, "/auth/admin/users/"+targetID.String()+"/promote", nil)
	req = withBearer(req)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d (body: %s)", rr.Code, rr.Body.String())
	}
	var body map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&body)
	if body["error"] != "user not found" {
		t.Errorf("error message: want %q, got %q", "user not found", body["error"])
	}
}

func TestAdminPromote_AlreadySuperAdmin_Returns409(t *testing.T) {
	targetID := uuid.New()
	superAdminRole := &entities.Role{ID: uuid.New(), Name: "SUPER_ADMIN"}
	userRepo := &authrepomocks.MockUserRepository{
		GetByIDUser: &entities.User{ID: targetID, Email: "x@example.com"},
	}
	roleRepo := &stubRoleRepoFull{
		rolesByName: map[string]*entities.Role{"SUPER_ADMIN": superAdminRole},
		userRoles: map[uuid.UUID][]*entities.Role{
			targetID: {superAdminRole},
		},
	}
	handler := buildAdminRouterWithTokenSvc(t, userRepo, roleRepo, superAdminClaims())

	req := httptest.NewRequest(http.MethodPost, "/auth/admin/users/"+targetID.String()+"/promote", nil)
	req = withBearer(req)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (body: %s)", rr.Code, rr.Body.String())
	}
	var body map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&body)
	if body["error"] != "user is already a super admin" {
		t.Errorf("error message: want %q, got %q", "user is already a super admin", body["error"])
	}
}

func TestAdminPromote_Success_Returns200(t *testing.T) {
	targetID := uuid.New()
	superAdminRole := &entities.Role{ID: uuid.New(), Name: "SUPER_ADMIN"}
	userRepo := &authrepomocks.MockUserRepository{
		GetByIDUser: &entities.User{ID: targetID, Email: "promoted@example.com"},
	}
	roleRepo := &stubRoleRepoFull{
		rolesByName: map[string]*entities.Role{"SUPER_ADMIN": superAdminRole},
		userRoles:   map[uuid.UUID][]*entities.Role{targetID: {}},
	}
	handler := buildAdminRouterWithTokenSvc(t, userRepo, roleRepo, superAdminClaims())

	req := httptest.NewRequest(http.MethodPost, "/auth/admin/users/"+targetID.String()+"/promote", nil)
	req = withBearer(req)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d (body: %s)", rr.Code, rr.Body.String())
	}
	var body map[string]interface{}
	_ = json.NewDecoder(rr.Body).Decode(&body)
	if body["id"] != targetID.String() {
		t.Errorf("id: want %q, got %v", targetID.String(), body["id"])
	}
	if body["isSuperAdmin"] != true {
		t.Errorf("isSuperAdmin: want true, got %v", body["isSuperAdmin"])
	}
}

func TestAdminPromote_BadUUID_Returns400(t *testing.T) {
	userRepo := &authrepomocks.MockUserRepository{}
	roleRepo := &stubRoleRepoFull{}
	handler := buildAdminRouterWithTokenSvc(t, userRepo, roleRepo, superAdminClaims())

	req := httptest.NewRequest(http.MethodPost, "/auth/admin/users/not-a-valid-uuid/promote", nil)
	req = withBearer(req)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d (body: %s)", rr.Code, rr.Body.String())
	}
}

// ── List users endpoint (DB nil → 503) ───────────────────────────────────────

func TestAdminListUsers_NoDB_Returns503(t *testing.T) {
	userRepo := &authrepomocks.MockUserRepository{}
	roleRepo := &stubRoleRepoFull{}
	handler := buildAdminRouterWithTokenSvc(t, userRepo, roleRepo, superAdminClaims())

	req := httptest.NewRequest(http.MethodGet, "/auth/admin/users", nil)
	req = withBearer(req)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 (no DB wired), got %d", rr.Code)
	}
}
