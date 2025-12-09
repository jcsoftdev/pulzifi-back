package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// TenantContextKey is the key used to store tenant in context
type contextKey string

const TenantContextKey contextKey = "tenant"

// TenantMiddleware extracts tenant from X-Tenant header or subdomain
// In monolith mode, tenant comes from X-Tenant header (set by load balancer)
// Format: X-Tenant: tenant_schema_name
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get tenant from X-Tenant header
		tenant := r.Header.Get("X-Tenant")

		// Fallback: extract from Host subdomain (optional)
		if tenant == "" {
			host := r.Host
			if parts := strings.Split(host, "."); len(parts) > 1 {
				tenant = parts[0]
			}
		}

	// Default to "jcsoftdev_inc" if not found
	if tenant == "" {
		tenant = "jcsoftdev_inc"
	}

	logger.Debug("Tenant extracted", zap.String("tenant", tenant), zap.String("path", r.URL.Path))

		// Store tenant in context
		ctx := context.WithValue(r.Context(), TenantContextKey, tenant)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTenantFromContext extracts tenant from request context
func GetTenantFromContext(ctx context.Context) string {
	tenant, ok := ctx.Value(TenantContextKey).(string)
	if !ok || tenant == "" {
		return "jcsoftdev_inc"
	}
	return tenant
}
