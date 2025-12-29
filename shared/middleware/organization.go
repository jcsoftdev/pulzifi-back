package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	authmw "github.com/jcsoftdev/pulzifi-back/modules/auth/infrastructure/middleware"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type OrganizationMiddleware struct {
	db *sql.DB
}

func NewOrganizationMiddleware(db *sql.DB) *OrganizationMiddleware {
	return &OrganizationMiddleware{db: db}
}

// RequireOrganizationMembership validates that the authenticated user belongs to the organization
func (m *OrganizationMiddleware) RequireOrganizationMembership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context (set by AuthMiddleware)
		userIDStr, ok := r.Context().Value(authmw.UserIDKey).(string)
		if !ok {
			logger.Warn("User ID not found in context for organization check")
			m.forbidden(w, "unauthorized")
			return
		}

		// Get subdomain from context (set by TenantMiddleware)
		subdomain := GetSubdomainFromContext(r.Context())
		if subdomain == "" {
			logger.Warn("Subdomain not found in context")
			m.forbidden(w, "tenant required")
			return
		}

		// Check if user belongs to organization
		isMember, err := m.checkMembership(r.Context(), userIDStr, subdomain)
		if err != nil {
			logger.Error("Failed to check organization membership",
				zap.String("user_id", userIDStr),
				zap.String("subdomain", subdomain),
				zap.Error(err))
			m.internalError(w, "failed to verify membership")
			return
		}

		if !isMember {
			logger.Warn("User does not belong to organization",
				zap.String("user_id", userIDStr),
				zap.String("subdomain", subdomain))
			m.forbidden(w, "access denied")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *OrganizationMiddleware) checkMembership(ctx context.Context, userID, subdomain string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 
			FROM public.organization_members om
			INNER JOIN public.organizations o ON om.organization_id = o.id
			WHERE om.user_id = $1::uuid 
			AND o.subdomain = $2
			AND o.deleted_at IS NULL
		)
	`

	var exists bool
	err := m.db.QueryRowContext(ctx, query, userID, subdomain).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (m *OrganizationMiddleware) forbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (m *OrganizationMiddleware) internalError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
