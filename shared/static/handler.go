package static

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// Setup configura el proxy o servicio de archivos estáticos según la configuración
// Debe ser llamado DESPUÉS de registrar todas las otras rutas
func Setup(router chi.Router, frontendURL, staticDir string, db *sql.DB, logger *zap.Logger) {
	if frontendURL != "" {
		// Modo desarrollo: proxy al servidor de frontend
		setupProxyNotFound(router, frontendURL, db, logger)
	} else if staticDir != "" {
		// Modo producción: servir archivos estáticos
		setupStaticNotFound(router, staticDir, logger)
	}
}

// subdomainExemptPrefixes are paths that must remain reachable from any
// host even when the subdomain doesn't resolve to an org. Only static
// assets and API namespaces are exempt — every app route (including
// /login, /register, /onboarding) should bounce invalid subdomains to
// the root domain so users land on a consistent host.
var subdomainExemptPrefixes = []string{
	"/_next",
	"/favicon",
	"/api",
}

func isSubdomainExempt(path string) bool {
	for _, p := range subdomainExemptPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// splitHostSubdomain returns (subdomain, baseHost) for a host like
// "acme.lvh.me" -> ("acme", "lvh.me"). For hosts without a subdomain
// (root domains, localhost, IPs) it returns ("", host).
func splitHostSubdomain(host string) (string, string) {
	hostOnly, port, err := net.SplitHostPort(host)
	if err != nil {
		hostOnly = host
		port = ""
	}
	if net.ParseIP(hostOnly) != nil || hostOnly == "localhost" {
		if port != "" {
			return "", net.JoinHostPort(hostOnly, port)
		}
		return "", hostOnly
	}
	parts := strings.Split(hostOnly, ".")
	if len(parts) < 3 {
		// e.g. lvh.me, pulzifi.com — no subdomain
		if port != "" {
			return "", net.JoinHostPort(hostOnly, port)
		}
		return "", hostOnly
	}
	subdomain := parts[0]
	base := strings.Join(parts[1:], ".")
	if port != "" {
		base = net.JoinHostPort(base, port)
	}
	return subdomain, base
}

// orgExistsForSubdomain returns true if a non-deleted organization exists
// with the given subdomain. Errors are treated as "exists" to fail open
// (avoid breaking the site if the DB hiccups during a page load).
func orgExistsForSubdomain(ctx context.Context, db *sql.DB, subdomain string) bool {
	if db == nil || subdomain == "" {
		return true
	}
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM public.organizations WHERE subdomain = $1 AND deleted_at IS NULL)`,
		subdomain,
	).Scan(&exists)
	if err != nil {
		return true
	}
	return exists
}

// setupProxyNotFound configures a reverse proxy for unmatched routes (404).
// All /api/* and /swagger/* paths return 404 (they should be Chi routes).
// Everything else is proxied to Next.js for page rendering.
func setupProxyNotFound(router chi.Router, frontendURL string, db *sql.DB, logger *zap.Logger) {
	target, err := url.Parse(frontendURL)
	if err != nil {
		logger.Error("Invalid frontend URL", zap.Error(err))
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.FlushInterval = -1 // SSE + HMR streaming
	proxy.Director = proxyDirector(proxy.Director)
	proxy.ErrorHandler = proxyErrorHandler(frontendURL, logger)

	logger.Info("Proxying unmatched routes to Next.js", zap.String("url", frontendURL))

	router.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Reserved Go API namespaces — anything else under /api/* (e.g. Payload CMS
		// REST endpoints at /api/users, /api/media, /api/globals/*) is proxied
		// through to Next.js where Payload mounts its handlers.
		if isReservedGoRoute(r.URL.Path) {
			http.NotFound(w, r)
			return
		}

		// Invalid-subdomain guard: if the host has a subdomain that doesn't
		// resolve to any org, redirect to the root domain so users land on
		// the marketing site instead of an orphan tenant URL. Exempt paths
		// (onboarding, login, etc.) must remain reachable from any host.
		if redirectInvalidSubdomain(w, r, db) {
			return
		}

		proxy.ServeHTTP(w, r)
	}))
}

// isReservedGoRoute reports whether the path belongs to a Go-handled namespace
// that should return 404 rather than being proxied to Next.js.
func isReservedGoRoute(path string) bool {
	return strings.HasPrefix(path, "/api/v1/") ||
		strings.HasPrefix(path, "/api/auth/") ||
		strings.HasPrefix(path, "/swagger")
}

// proxyDirector wraps the default reverse-proxy director to forward host headers
// and inject the X-Tenant header derived from the request subdomain.
func proxyDirector(originalDirector func(*http.Request)) func(*http.Request) {
	return func(req *http.Request) {
		// Capture original host BEFORE the default director rewrites the URL
		origHost := req.Host

		originalDirector(req)

		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		req.Header.Set("X-Forwarded-Host", origHost)

		proto := "http"
		if req.TLS != nil {
			proto = "https"
		}
		if fp := req.Header.Get("X-Forwarded-Proto"); fp != "" {
			proto = fp
		}
		req.Header.Set("X-Forwarded-Proto", proto)

		// Extract subdomain and add as X-Tenant header (*.localhost and *.app.local)
		host := strings.Split(origHost, ":")[0]
		if strings.Contains(host, ".localhost") || strings.HasSuffix(host, ".app.local") || strings.HasSuffix(host, ".local") {
			parts := strings.Split(host, ".")
			if len(parts) >= 2 {
				tenant := parts[0]
				if tenant != "" && tenant != "localhost" && tenant != "app" {
					req.Header.Set("X-Tenant", tenant)
				}
			}
		}
	}
}

// proxyErrorHandler returns the reverse-proxy error handler. Client cancellations
// are logged at debug; everything else is logged as an error and returns 502.
func proxyErrorHandler(frontendURL string, logger *zap.Logger) func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		if errors.Is(err, context.Canceled) {
			logger.Debug("Proxy request cancelled by client", zap.String("path", r.URL.Path))
			return
		}
		logger.Error("Proxy error", zap.Error(err), zap.String("url", frontendURL), zap.String("path", r.URL.Path))
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("Frontend server unavailable"))
	}
}

// redirectInvalidSubdomain redirects to the root domain when the request host has
// a subdomain that doesn't resolve to any org. Returns true if a redirect was written.
func redirectInvalidSubdomain(w http.ResponseWriter, r *http.Request, db *sql.DB) bool {
	if isSubdomainExempt(r.URL.Path) {
		return false
	}
	subdomain, base := splitHostSubdomain(r.Host)
	if subdomain == "" || orgExistsForSubdomain(r.Context(), db, subdomain) {
		return false
	}
	proto := r.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}
	target := proto + "://" + base + r.URL.RequestURI()
	http.Redirect(w, r, target, http.StatusTemporaryRedirect)
	return true
}

// setupStaticNotFound configura archivos estáticos para rutas no encontradas
func setupStaticNotFound(router chi.Router, staticDir string, logger *zap.Logger) {
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		logger.Warn("Static directory not found", zap.String("dir", staticDir))
		return
	}

	logger.Info("Serving static files", zap.String("dir", staticDir))
	fileServer := http.FileServer(http.Dir(staticDir))

	router.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No servir archivos para rutas de API o Swagger
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/swagger") {
			http.NotFound(w, r)
			return
		}
		fullPath := filepath.Join(staticDir, r.URL.Path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			// Archivo no existe, servir index.html (SPA)
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}
		fileServer.ServeHTTP(w, r)
	}))
}
