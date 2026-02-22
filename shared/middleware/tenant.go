package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// validSchemaName matches only safe identifier characters (alphanumeric + underscore).
var validSchemaName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

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
// 4. Allows certain paths to bypass tenant requirement (e.g., /swagger, /health)
func TenantMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow certain paths without tenant
			if isPublicPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			subdomain := extractSubdomain(r)

			// For development (localhost/app.local), allow accessing without specific tenant
			// In production, all requests should have a valid subdomain
			if !isValidSubdomain(subdomain) {
				logger.Warn("No valid subdomain in request",
					zap.String("host", r.Host),
					zap.String("subdomain", subdomain),
					zap.String("x-tenant", r.Header.Get("X-Tenant")))
				http.Error(w, "Subdomain is required", http.StatusBadRequest)
				return
			}

			// Development: if subdomain is generic (app, localhost), skip schema resolution
			// Production: always resolve to schema
			if isGenericDomain(subdomain) {
				logger.Debug("Generic domain (development mode), skipping schema resolution",
					zap.String("subdomain", subdomain),
					zap.String("host", r.Host))
				// For development, use subdomain as schema too (will be handled by frontend proxy)
				ctx := buildContextWithTenant(r.Context(), subdomain, subdomain)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Production: resolve subdomain to actual schema
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

func isPublicPath(path string) bool {
	publicPaths := []string{
		"/api/v1/swagger",
		"/api/v1/health",
		"/api/v1/auth/login",     // Login no requiere tenant
		"/api/v1/auth/register",         // Register no requiere tenant
		"/api/v1/auth/check-subdomain",  // Subdomain check no requiere tenant
		"/api/v1/auth/me",        // Current user (uses JWT, not tenant)
		"/api/v1/auth/refresh",   // Token refresh
		"/api/v1/auth/providers", // OAuth providers
		"/api/v1/auth/csrf",      // CSRF token
		"/api/docs",              // API Documentation
		"/swagger",
		"/health",
		"/docs",
	}

	for _, prefix := range publicPaths {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func extractSubdomain(r *http.Request) string {
	// Try X-Tenant header first (set by proxy)
	if subdomain := r.Header.Get("X-Tenant"); subdomain != "" {
		return subdomain
	}

	// Fallback: try X-Forwarded-Host (passed by Node.js server-side requests)
	if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host := strings.Split(forwardedHost, ":")[0]
		parts := strings.Split(host, ".")
		if len(parts) > 1 {
			return parts[0]
		}
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

func isGenericDomain(subdomain string) bool {
	// Generic domains used in development (app.local, app, etc)
	// These don't map to specific organizations
	genericDomains := []string{"app", "localhost", "127.0.0.1"}
	for _, domain := range genericDomains {
		if subdomain == domain {
			return true
		}
	}
	return false
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

// GetSetSearchPathSQL returns the proper SQL command to set the search_path.
// The tenant (schema) name is validated to contain only safe identifier characters.
// Returns a safe SET command or a no-op if the name is invalid.
func GetSetSearchPathSQL(tenant string) string {
	if !validSchemaName.MatchString(tenant) {
		logger.Warn("Invalid schema name rejected", zap.String("tenant", tenant))
		// Return a safe no-op that changes nothing
		return "SELECT 1"
	}
	return `SET search_path TO "` + tenant + `", public`
}

// GetTenantFromContextOrError extracts tenant from context and returns error if not found
func GetTenantFromContextOrError(ctx context.Context) (string, error) {
	tenant := GetTenantFromContext(ctx)
	if tenant == "" {
		return "", fmt.Errorf("tenant not found in context")
	}
	return tenant, nil
}

// RequireTenant is a middleware that ensures tenant is present in context
func RequireTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenant := GetTenantFromContext(r.Context())
		if tenant == "" {
			logger.Warn("Tenant not found in context",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
			http.Error(w, "Tenant not found", http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, r)
	})
}
