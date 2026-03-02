package static

import (
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
func Setup(router chi.Router, frontendURL, staticDir string, logger *zap.Logger) {
	if frontendURL != "" {
		// Modo desarrollo: proxy al servidor de frontend
		setupProxyNotFound(router, frontendURL, logger)
	} else if staticDir != "" {
		// Modo producción: servir archivos estáticos
		setupStaticNotFound(router, staticDir, logger)
	}
}

// setupProxyNotFound configures a reverse proxy for unmatched routes (404).
// All /api/* and /swagger/* paths return 404 (they should be Chi routes).
// Everything else is proxied to Next.js for page rendering.
func setupProxyNotFound(router chi.Router, frontendURL string, logger *zap.Logger) {
	target, err := url.Parse(frontendURL)
	if err != nil {
		logger.Error("Invalid frontend URL", zap.Error(err))
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.FlushInterval = -1 // SSE + HMR streaming

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		// Capture original host BEFORE the default director rewrites the URL
		origHost := req.Host

		originalDirector(req)

		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		req.Header.Set("X-Forwarded-Host", origHost)

		proto := "http"
		if fp := req.Header.Get("X-Forwarded-Proto"); fp != "" {
			proto = fp
		}
		req.Header.Set("X-Forwarded-Proto", proto)

		// Extract subdomain and add as X-Tenant header
		if strings.Contains(origHost, ".localhost") {
			parts := strings.Split(origHost, ".")
			if len(parts) >= 2 {
				tenant := strings.Split(parts[0], ":")[0]
				if tenant != "" && tenant != "localhost" {
					req.Header.Set("X-Tenant", tenant)
				}
			}
		}
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("Proxy error", zap.Error(err), zap.String("url", frontendURL), zap.String("path", r.URL.Path))
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Frontend server unavailable"))
	}

	logger.Info("Proxying unmatched routes to Next.js", zap.String("url", frontendURL))

	router.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// All /api/* paths should be registered Chi routes — return 404
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/swagger") {
			http.NotFound(w, r)
			return
		}
		proxy.ServeHTTP(w, r)
	}))
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

// setupProxy DEPRECATED - usa setupProxyNotFound en su lugar
func setupProxy(router chi.Router, frontendURL string, logger *zap.Logger) {
	target, err := url.Parse(frontendURL)
	if err != nil {
		logger.Error("Invalid frontend URL", zap.Error(err))
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Modificar Director para preservar headers importantes
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		// Preservar headers originales
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		req.Header.Set("X-Forwarded-Proto", "http")
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("Proxy error", zap.Error(err), zap.String("url", frontendURL), zap.String("path", r.URL.Path))
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Frontend server unavailable"))
	}

	logger.Info("Proxying to frontend dev server", zap.String("url", frontendURL))

	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
			http.NotFound(w, r)
			return
		}
		logger.Debug("Proxying request", zap.String("path", r.URL.Path), zap.String("method", r.Method))
		proxy.ServeHTTP(w, r)
	})
}

// setupStatic configura el servidor de archivos estáticos (producción)
func setupStatic(router chi.Router, staticDir string, logger *zap.Logger) {
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		logger.Warn("Static directory not found", zap.String("dir", staticDir))
		return
	}

	logger.Info("Serving static files", zap.String("dir", staticDir))
	fileServer := http.FileServer(http.Dir(staticDir))

	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") {
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
	})
}
