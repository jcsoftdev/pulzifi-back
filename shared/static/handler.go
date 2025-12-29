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

// setupProxyNotFound configura el proxy solo para rutas no encontradas (404)
func setupProxyNotFound(router chi.Router, frontendURL string, logger *zap.Logger) {
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
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		req.Header.Set("X-Forwarded-Proto", "http")

		// Extract subdomain and add as X-Tenant header
		host := req.Host
		if strings.Contains(host, ".localhost") {
			// Extract subdomain from host like "tenant.localhost:9090"
			parts := strings.Split(host, ".")
			if len(parts) >= 2 {
				tenant := parts[0]
				// Remove port if present
				tenant = strings.Split(tenant, ":")[0]
				if tenant != "" && tenant != "localhost" {
					req.Header.Set("X-Tenant", tenant)
					logger.Debug("Extracted tenant from subdomain", zap.String("tenant", tenant), zap.String("host", host))
				}
			}
		}
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("Proxy error", zap.Error(err), zap.String("url", frontendURL), zap.String("path", r.URL.Path))
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Frontend server unavailable"))
	}

	logger.Info("Proxying to frontend dev server", zap.String("url", frontendURL))

	// Usar NotFound para solo proxear rutas no encontradas
	router.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No proxear rutas de API (excepto /api/auth que es NextAuth)
		if strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/api/auth") {
			http.NotFound(w, r)
			return
		}
		// No proxear Swagger
		if strings.HasPrefix(r.URL.Path, "/swagger") {
			http.NotFound(w, r)
			return
		}
		logger.Debug("Proxying request via NotFound", zap.String("path", r.URL.Path), zap.String("method", r.Method))
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
