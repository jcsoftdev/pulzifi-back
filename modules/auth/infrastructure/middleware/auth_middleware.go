package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
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
	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Warn("Missing authorization header")
			m.unauthorized(w, "missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("Invalid authorization header format", zap.String("header", authHeader[:20]))
			m.unauthorized(w, "invalid authorization header format")
			return
		}

		token := parts[1]
		logger.Info("Attempting token validation", zap.String("token_preview", token[:30]))

		claims, err := m.tokenService.ValidateToken(r.Context(), token)
		if err != nil {
			logger.Error("Token validation failed", zap.Error(err), zap.String("token_preview", token[:30]))
			m.unauthorized(w, "invalid or expired token")
			return
		}

		logger.Info("Token validated successfully", zap.String("user_id", claims.UserID.String()))

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, claims.UserID.String())
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, UserRolesKey, claims.Roles)
		ctx = context.WithValue(ctx, UserPermsKey, claims.Permissions)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
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
			hasPermission := false
			for _, perm := range permissions {
				if perm == requiredPerm {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				m.forbidden(w, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
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

			hasRole := false
			for _, role := range roles {
				if role == requiredRole {
					hasRole = true
					break
				}
			}

			if !hasRole {
				m.forbidden(w, "insufficient role")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *AuthMiddleware) unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func (m *AuthMiddleware) forbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
