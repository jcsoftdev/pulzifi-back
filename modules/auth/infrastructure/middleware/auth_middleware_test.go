package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/cookies"
)

// --- Mock TokenService ---

type mockTokenService struct {
	validateFn func(ctx context.Context, token string) (*services.TokenClaims, error)
}

func (m *mockTokenService) GenerateAccessToken(_ context.Context, _ uuid.UUID, _ string) (string, error) {
	return "", nil
}

func (m *mockTokenService) GenerateRefreshToken(_ context.Context, _ uuid.UUID) (string, error) {
	return "", nil
}

func (m *mockTokenService) ValidateToken(ctx context.Context, token string) (*services.TokenClaims, error) {
	return m.validateFn(ctx, token)
}

func (m *mockTokenService) GetTokenExpiration() time.Duration {
	return 15 * time.Minute
}

func (m *mockTokenService) GetRefreshTokenExpiration() time.Time {
	return time.Now().Add(7 * 24 * time.Hour)
}

// --- Helpers ---

// newRequestWithAccessToken creates an http.Request with an access_token cookie.
func newRequestWithAccessToken(token string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  cookies.AccessTokenCookie,
		Value: token,
	})
	return req
}

// successHandler is a simple handler that writes 200 OK.
func successHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// --- Tests ---

func TestAuthenticate(t *testing.T) {
	validUserID := uuid.New()

	tests := []struct {
		name           string
		request        *http.Request
		validateFn     func(ctx context.Context, token string) (*services.TokenClaims, error)
		wantStatus     int
		checkCtxValues bool
		wantUserID     string
		wantEmail      string
		wantRoles      []string
		wantPerms      []string
	}{
		{
			name:    "valid token sets context values",
			request: newRequestWithAccessToken("valid-token"),
			validateFn: func(_ context.Context, _ string) (*services.TokenClaims, error) {
				return &services.TokenClaims{
					UserID:      validUserID,
					Email:       "alice@example.com",
					Roles:       []string{"admin", "user"},
					Permissions: []string{"read:users", "write:users"},
				}, nil
			},
			wantStatus:     http.StatusOK,
			checkCtxValues: true,
			wantUserID:     validUserID.String(),
			wantEmail:      "alice@example.com",
			wantRoles:      []string{"admin", "user"},
			wantPerms:      []string{"read:users", "write:users"},
		},
		{
			name:    "missing token cookie returns 401",
			request: httptest.NewRequest(http.MethodGet, "/", nil), // no cookie
			validateFn: func(_ context.Context, _ string) (*services.TokenClaims, error) {
				t.Fatal("ValidateToken should not be called when cookie is missing")
				return nil, nil
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "invalid token returns 401",
			request: newRequestWithAccessToken("bad-token"),
			validateFn: func(_ context.Context, _ string) (*services.TokenClaims, error) {
				return nil, errors.New("token expired")
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "empty token value returns 401",
			request: newRequestWithAccessToken(""),
			validateFn: func(_ context.Context, _ string) (*services.TokenClaims, error) {
				t.Fatal("ValidateToken should not be called for empty token")
				return nil, nil
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &mockTokenService{validateFn: tt.validateFn}
			mw := NewAuthMiddleware(ts)

			var capturedCtx context.Context
			inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedCtx = r.Context()
				w.WriteHeader(http.StatusOK)
			})

			rr := httptest.NewRecorder()
			mw.Authenticate(inner).ServeHTTP(rr, tt.request)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}

			if tt.checkCtxValues && capturedCtx != nil {
				gotUserID, _ := capturedCtx.Value(UserIDKey).(string)
				if gotUserID != tt.wantUserID {
					t.Errorf("expected user_id %q, got %q", tt.wantUserID, gotUserID)
				}
				gotEmail, _ := capturedCtx.Value(UserEmailKey).(string)
				if gotEmail != tt.wantEmail {
					t.Errorf("expected email %q, got %q", tt.wantEmail, gotEmail)
				}
				gotRoles, _ := capturedCtx.Value(UserRolesKey).([]string)
				if len(gotRoles) != len(tt.wantRoles) {
					t.Errorf("expected %d roles, got %d", len(tt.wantRoles), len(gotRoles))
				}
				gotPerms, _ := capturedCtx.Value(UserPermsKey).([]string)
				if len(gotPerms) != len(tt.wantPerms) {
					t.Errorf("expected %d permissions, got %d", len(tt.wantPerms), len(gotPerms))
				}
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name         string
		ctxRoles     interface{} // value set in context for UserRolesKey
		requiredRole string
		wantStatus   int
	}{
		{
			name:         "user has required role",
			ctxRoles:     []string{"admin", "user"},
			requiredRole: "admin",
			wantStatus:   http.StatusOK,
		},
		{
			name:         "user lacks required role",
			ctxRoles:     []string{"user"},
			requiredRole: "admin",
			wantStatus:   http.StatusForbidden,
		},
		{
			name:         "no roles in context",
			ctxRoles:     nil,
			requiredRole: "admin",
			wantStatus:   http.StatusForbidden,
		},
		{
			name:         "wrong type in context",
			ctxRoles:     "admin", // string instead of []string
			requiredRole: "admin",
			wantStatus:   http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &mockTokenService{}
			mw := NewAuthMiddleware(ts)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.ctxRoles != nil {
				ctx := context.WithValue(req.Context(), UserRolesKey, tt.ctxRoles)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			mw.RequireRole(tt.requiredRole)(successHandler()).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}

func TestRequirePermission(t *testing.T) {
	tests := []struct {
		name       string
		ctxPerms   interface{}
		resource   string
		action     string
		wantStatus int
	}{
		{
			name:       "user has required permission",
			ctxPerms:   []string{"users:read", "users:write"},
			resource:   "users",
			action:     "read",
			wantStatus: http.StatusOK,
		},
		{
			name:       "user lacks required permission",
			ctxPerms:   []string{"users:read"},
			resource:   "users",
			action:     "write",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "no permissions in context",
			ctxPerms:   nil,
			resource:   "users",
			action:     "read",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "wrong type in context",
			ctxPerms:   "users:read",
			resource:   "users",
			action:     "read",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "partial match does not grant access",
			ctxPerms:   []string{"users:readonly"},
			resource:   "users",
			action:     "read",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := &mockTokenService{}
			mw := NewAuthMiddleware(ts)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.ctxPerms != nil {
				ctx := context.WithValue(req.Context(), UserPermsKey, tt.ctxPerms)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			mw.RequirePermission(tt.resource, tt.action)(successHandler()).ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}
