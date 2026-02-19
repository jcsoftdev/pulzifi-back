package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/cookies"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type ContextKey string

const (
	UserIDKey    ContextKey = "user_id"
	UserEmailKey ContextKey = "user_email"
	UserRolesKey ContextKey = "user_roles"
	UserPermsKey ContextKey = "user_permissions"
)

type AuthMiddleware struct {
	tokenService services.TokenService
}

func NewAuthMiddleware(tokenService services.TokenService) *AuthMiddleware {
	return &AuthMiddleware{tokenService: tokenService}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, userEmail, roles, permissions, ok := m.resolveAuthContext(r)
		if !ok {
			m.unauthorized(w, "unauthorized")
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, userID)
		ctx = context.WithValue(ctx, UserEmailKey, userEmail)
		ctx = context.WithValue(ctx, UserRolesKey, roles)
		ctx = context.WithValue(ctx, UserPermsKey, permissions)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) resolveAuthContext(r *http.Request) (string, string, []string, []string, bool) {
	tokenStr, err := cookies.GetTokenFromCookie(r, cookies.AccessTokenCookie)
	if err != nil || tokenStr == "" {
		logger.Warn("Missing access token cookie")
		return "", "", nil, nil, false
	}

	claims, err := m.tokenService.ValidateToken(r.Context(), tokenStr)
	if err != nil {
		logger.Warn("Invalid access token", zap.Error(err))
		return "", "", nil, nil, false
	}

	return claims.UserID.String(), claims.Email, claims.Roles, claims.Permissions, true
}

func (m *AuthMiddleware) RequirePermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions, ok := r.Context().Value(UserPermsKey).([]string)
			if !ok {
				m.forbidden(w, "no permissions found")
				return
			}

			requiredPerm := resource + ":" + action
			for _, perm := range permissions {
				if perm == requiredPerm {
					next.ServeHTTP(w, r)
					return
				}
			}

			m.forbidden(w, "insufficient permissions")
		})
	}
}

func (m *AuthMiddleware) RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles, ok := r.Context().Value(UserRolesKey).([]string)
			if !ok {
				m.forbidden(w, "no roles found")
				return
			}

			for _, role := range roles {
				if role == requiredRole {
					next.ServeHTTP(w, r)
					return
				}
			}

			m.forbidden(w, "insufficient role")
		})
	}
}

// ValidateTokenFromRequest validates the access token cookie and returns the user ID.
func (m *AuthMiddleware) ValidateTokenFromRequest(r *http.Request) (uuid.UUID, error) {
	tokenStr, err := cookies.GetTokenFromCookie(r, cookies.AccessTokenCookie)
	if err != nil || tokenStr == "" {
		return uuid.Nil, errors.New("missing access token")
	}

	claims, err := m.tokenService.ValidateToken(r.Context(), tokenStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid or expired token")
	}

	return claims.UserID, nil
}

func (m *AuthMiddleware) unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (m *AuthMiddleware) forbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
