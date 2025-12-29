package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// TenantContextKey is the key used to store tenant schema in context
type contextKey string

const (
	TenantContextKey    contextKey = "tenant"
	SubdomainContextKey contextKey = "subdomain"
)

// TenantMiddleware extracts subdomain and resolves it to a tenant schema
// 1. Extracts subdomain from X-Tenant header or Host
// 2. Queries public.organizations to get the schema_name
// 3. Stores schema_name in context
func TenantMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			subdomain := extractSubdomain(r)
			if !isValidSubdomain(subdomain) {
				logger.Warn("No subdomain provided in request", zap.String("host", r.Host))
				http.Error(w, "Subdomain is required", http.StatusBadRequest)
				return
			}

			schemaName, err := resolveSchema(db, subdomain)
			if err != nil {
				handleSchemaResolutionError(w, subdomain, err)
				return
			}

			logger.Debug("Tenant resolved",
				zap.String("subdomain", subdomain),
				zap.String("schema", schemaName),
				zap.String("path", r.URL.Path))

			ctx := buildContextWithTenant(r.Context(), subdomain, schemaName)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractSubdomain(r *http.Request) string {
	// Try X-Tenant header first (set by proxy)
	if subdomain := r.Header.Get("X-Tenant"); subdomain != "" {
		return subdomain
	}

	// Fallback: extract from Host
	host := strings.Split(r.Host, ":")[0]
	parts := strings.Split(host, ".")
	if len(parts) > 1 {
		return parts[0]
	}

	return ""
}

func isValidSubdomain(subdomain string) bool {
	return subdomain != "" && subdomain != "localhost"
}

func resolveSchema(db *sql.DB, subdomain string) (string, error) {
	query := `SELECT schema_name FROM public.organizations WHERE subdomain = $1 AND deleted_at IS NULL LIMIT 1`
	var schemaName string
	err := db.QueryRow(query, subdomain).Scan(&schemaName)
	return schemaName, err
}

func handleSchemaResolutionError(w http.ResponseWriter, subdomain string, err error) {
	if err == sql.ErrNoRows {
		logger.Warn("Organization not found for subdomain", zap.String("subdomain", subdomain))
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	logger.Error("Failed to resolve subdomain to schema",
		zap.String("subdomain", subdomain),
		zap.Error(err))
	http.Error(w, "Internal server error", http.StatusInternalServerError)
}

func buildContextWithTenant(ctx context.Context, subdomain, schemaName string) context.Context {
	ctx = context.WithValue(ctx, SubdomainContextKey, subdomain)
	ctx = context.WithValue(ctx, TenantContextKey, schemaName)
	return ctx
}

// GetTenantFromContext extracts tenant schema from request context
func GetTenantFromContext(ctx context.Context) string {
	tenant, ok := ctx.Value(TenantContextKey).(string)
	if !ok || tenant == "" {
		return ""
	}
	return tenant
}

// GetSubdomainFromContext extracts subdomain from request context
func GetSubdomainFromContext(ctx context.Context) string {
	subdomain, ok := ctx.Value(SubdomainContextKey).(string)
	if !ok || subdomain == "" {
		return ""
	}
	return subdomain
}
