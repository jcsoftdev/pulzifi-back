package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
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
	sessionRepo repositories.SessionRepository
	userRepo    repositories.UserRepository
	roleRepo    repositories.RoleRepository
	permRepo    repositories.PermissionRepository
}

func NewAuthMiddleware(
	sessionRepo repositories.SessionRepository,
	userRepo repositories.UserRepository,
	roleRepo repositories.RoleRepository,
	permRepo repositories.PermissionRepository,
) *AuthMiddleware {
	return &AuthMiddleware{
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
		roleRepo:    roleRepo,
		permRepo:    permRepo,
	}
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
	sessionID, err := cookies.GetSessionIDFromCookie(r)
	if err != nil || sessionID == "" {
		logger.Warn("Missing session cookie")
		return "", "", nil, nil, false
	}

	session, err := m.sessionRepo.FindByID(r.Context(), sessionID)
	if err != nil || session == nil || session.IsExpired() {
		if session != nil && session.IsExpired() {
			_ = m.sessionRepo.DeleteByID(r.Context(), sessionID)
		}
		logger.Warn("Session validation failed", zap.Error(err))
		return "", "", nil, nil, false
	}

	user, err := m.userRepo.GetByID(r.Context(), session.UserID)
	if err != nil || user == nil {
		logger.Warn("User not found for session", zap.Error(err), zap.String("user_id", session.UserID.String()))
		return "", "", nil, nil, false
	}

	roles, err := m.roleRepo.GetUserRoles(r.Context(), session.UserID)
	if err != nil {
		logger.Error("Failed to load user roles", zap.Error(err))
		return "", "", nil, nil, false
	}

	permissions, err := m.permRepo.GetUserPermissions(r.Context(), session.UserID)
	if err != nil {
		logger.Error("Failed to load user permissions", zap.Error(err))
		return "", "", nil, nil, false
	}

	roleNames := make([]string, 0, len(roles))
	for _, role := range roles {
		roleNames = append(roleNames, role.Name)
	}

	permissionNames := make([]string, 0, len(permissions))
	for _, perm := range permissions {
		permissionNames = append(permissionNames, perm.Name)
	}

	logger.Info("Session validated successfully", zap.String("user_id", session.UserID.String()))
	return session.UserID.String(), user.Email, roleNames, permissionNames, true
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

func (m *AuthMiddleware) SessionIDFromRequest(r *http.Request) (string, error) {
	return cookies.GetSessionIDFromCookie(r)
}

func (m *AuthMiddleware) ValidateSessionFromRequest(r *http.Request) (uuid.UUID, error) {
	sessionID, err := cookies.GetSessionIDFromCookie(r)
	if err != nil || sessionID == "" {
		return uuid.Nil, errors.New("missing session")
	}

	session, err := m.sessionRepo.FindByID(r.Context(), sessionID)
	if err != nil || session == nil || session.IsExpired() {
		return uuid.Nil, errors.New("invalid or expired session")
	}

	return session.UserID, nil
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
