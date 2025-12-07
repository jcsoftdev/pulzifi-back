package http

import (
	"net/http"
	"time"

	"github.com/go-chi/httplog/v2"
	"go.uber.org/zap"
)

// ChiMiddleware provides common Chi middleware setup
type ChiMiddleware struct {
	logger *zap.Logger
}

// NewChiMiddleware creates a new Chi middleware helper
func NewChiMiddleware(logger *zap.Logger) *ChiMiddleware {
	return &ChiMiddleware{
		logger: logger,
	}
}

// Logging returns a logging middleware using httplog
func (cm *ChiMiddleware) Logging() func(next http.Handler) http.Handler {
	return httplog.RequestLogger(httplog.NewLogger("api"))
}

// Recovery returns a panic recovery middleware
func (cm *ChiMiddleware) Recovery() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					cm.logger.Error("panic recovered", zap.Any("panic", err))
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID returns a middleware that adds a request ID to the context
func (cm *ChiMiddleware) RequestID() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}
			w.Header().Set("X-Request-ID", requestID)
			next.ServeHTTP(w, r)
		})
	}
}

// Timeout returns a middleware that times out requests
func (cm *ChiMiddleware) Timeout(duration time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, duration, "Request timeout")
	}
}

// CORS returns a middleware for CORS
func (cm *ChiMiddleware) CORS(allowedOrigins []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			for _, allowed := range allowedOrigins {
				if origin == allowed || allowed == "*" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
					w.Header().Set("Access-Control-Allow-Credentials", "true")
					break
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ContentType returns a middleware that sets content type
func (cm *ChiMiddleware) ContentType(contentType string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)
			next.ServeHTTP(w, r)
		})
	}
}

// Helper function to generate request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + string(rune(int64(time.Now().Nanosecond())))
}
