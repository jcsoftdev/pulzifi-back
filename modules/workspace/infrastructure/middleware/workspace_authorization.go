package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/value_objects"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	sharedmw "github.com/jcsoftdev/pulzifi-back/shared/middleware"
	"go.uber.org/zap"
)

type ContextKey string

const (
	WorkspaceRoleKey ContextKey = "workspace_role"
)

// HTTP constants
const (
	contentTypeHeader = "Content-Type"
	applicationJSON   = "application/json"
)

// Error messages
const (
	errWorkspaceIDRequired    = "workspace ID is required"
	errInvalidWorkspaceIDFmt  = "invalid workspace ID format"
	errUserNotAuthenticated   = "user not authenticated"
	errInvalidUserID          = "invalid user ID"
	errTenantNotFound         = "tenant not found"
	errNotWorkspaceMember     = "you are not a member of this workspace"
	errMembershipVerification = "failed to verify workspace membership"
	errInvalidWorkspaceRole   = "invalid workspace role"
	errWorkspaceRoleNotFound  = "workspace role not found - membership required"
	errInsufficientRole       = "insufficient permissions in this workspace"
)

// WorkspaceAuthorizationMiddleware provides two-level authorization:
// Level 1 (Global): User has permission to use workspaces feature (via Auth middleware)
// Level 2 (Domain): User has specific role in the requested workspace (via this middleware)
type WorkspaceAuthorizationMiddleware struct {
	db *sql.DB
}

func NewWorkspaceAuthorizationMiddleware(db *sql.DB) *WorkspaceAuthorizationMiddleware {
	return &WorkspaceAuthorizationMiddleware{db: db}
}

// RequireWorkspaceMembership ensures user is a member of the workspace
// Note: This is Level 2 authorization (domain-specific)
// Level 1 (global permission check) should already be done by RequirePermission middleware
func (m *WorkspaceAuthorizationMiddleware) RequireWorkspaceMembership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		workspaceUUID, userID, tenant, err := m.extractRequestParams(w, r)
		if err != nil {
			return // Error already handled in extractRequestParams
		}

		role, err := m.getMemberRole(r.Context(), tenant, workspaceUUID, userID)
		if err != nil {
			m.handleMembershipError(w, err, workspaceUUID.String(), userID.String())
			return
		}

		// Add workspace role to context for use in handlers
		ctx := context.WithValue(r.Context(), WorkspaceRoleKey, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractRequestParams extracts and validates workspace ID, user ID, and tenant from request
func (m *WorkspaceAuthorizationMiddleware) extractRequestParams(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, string, error) {
	workspaceID := chi.URLParam(r, "id")
	if workspaceID == "" {
		m.badRequest(w, errWorkspaceIDRequired)
		return uuid.Nil, uuid.Nil, "", errors.New("missing workspace ID")
	}

	workspaceUUID, err := uuid.Parse(workspaceID)
	if err != nil {
		m.badRequest(w, errInvalidWorkspaceIDFmt)
		return uuid.Nil, uuid.Nil, "", err
	}

	userIDStr, ok := r.Context().Value(authmw.UserIDKey).(string)
	if !ok {
		m.unauthorized(w, errUserNotAuthenticated)
		return uuid.Nil, uuid.Nil, "", errors.New("user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		m.unauthorized(w, errInvalidUserID)
		return uuid.Nil, uuid.Nil, "", err
	}

	// Get tenant schema from context (set by TenantMiddleware)
	tenant := sharedmw.GetTenantFromContext(r.Context())
	if tenant == "" {
		m.badRequest(w, errTenantNotFound)
		return uuid.Nil, uuid.Nil, "", errors.New("tenant not found")
	}

	return workspaceUUID, userID, tenant, nil
}

// getMemberRole retrieves the workspace member role from database
func (m *WorkspaceAuthorizationMiddleware) getMemberRole(ctx context.Context, tenant string, workspaceID, userID uuid.UUID) (value_objects.WorkspaceRole, error) {
	query := `
		SELECT role
		FROM ` + tenant + `.workspace_members
		WHERE workspace_id = $1 AND user_id = $2 AND removed_at IS NULL
	`

	var roleStr string
	err := m.db.QueryRowContext(ctx, query, workspaceID, userID).Scan(&roleStr)
	if err != nil {
		return "", err
	}

	return value_objects.NewWorkspaceRole(roleStr)
}

// handleMembershipError logs and returns appropriate error response
func (m *WorkspaceAuthorizationMiddleware) handleMembershipError(w http.ResponseWriter, err error, workspaceID, userID string) {
	if err == sql.ErrNoRows {
		logger.Warn("User is not a member of workspace",
			zap.String("workspace_id", workspaceID),
			zap.String("user_id", userID),
		)
		m.forbidden(w, errNotWorkspaceMember)
		return
	}

	logger.Error("Failed to get workspace member",
		zap.Error(err),
		zap.String("workspace_id", workspaceID),
		zap.String("user_id", userID),
	)
	m.internalError(w, errMembershipVerification)
}

// RequireWorkspaceRole ensures user has a specific role in the workspace
// Must be used after RequireWorkspaceMembership
func (m *WorkspaceAuthorizationMiddleware) RequireWorkspaceRole(minimumRole value_objects.WorkspaceRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(WorkspaceRoleKey).(value_objects.WorkspaceRole)
			if !ok {
				m.forbidden(w, errWorkspaceRoleNotFound)
				return
			}

			// Check if user has sufficient role
			if !m.hasMinimumRole(role, minimumRole) {
				logger.Warn("Insufficient workspace role",
					zap.String("required", minimumRole.String()),
					zap.String("actual", role.String()),
				)
				m.forbidden(w, errInsufficientRole)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// hasMinimumRole checks if user's role meets the minimum required role
func (m *WorkspaceAuthorizationMiddleware) hasMinimumRole(userRole, requiredRole value_objects.WorkspaceRole) bool {
	// Role hierarchy: owner > editor > viewer
	roleHierarchy := map[value_objects.WorkspaceRole]int{
		value_objects.RoleOwner:  3,
		value_objects.RoleEditor: 2,
		value_objects.RoleViewer: 1,
	}

	return roleHierarchy[userRole] >= roleHierarchy[requiredRole]
}

func (m *WorkspaceAuthorizationMiddleware) unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set(contentTypeHeader, applicationJSON)
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func (m *WorkspaceAuthorizationMiddleware) forbidden(w http.ResponseWriter, message string) {
	w.Header().Set(contentTypeHeader, applicationJSON)
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func (m *WorkspaceAuthorizationMiddleware) badRequest(w http.ResponseWriter, message string) {
	w.Header().Set(contentTypeHeader, applicationJSON)
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func (m *WorkspaceAuthorizationMiddleware) internalError(w http.ResponseWriter, message string) {
	w.Header().Set(contentTypeHeader, applicationJSON)
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
