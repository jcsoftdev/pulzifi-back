package middleware

import (
	"net/http"
	"time"

	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// ResponseLoggerMiddleware logs all HTTP responses with status codes
func ResponseLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		// Log the response
		fields := []zap.Field{
			zap.String("method", r.Method),
			zap.String("path", r.RequestURI),
			zap.Int("status", wrapped.statusCode),
			zap.Duration("duration", duration),
		}

		// Add tenant if available
		if tenant := GetTenantFromContext(r.Context()); tenant != "" {
			fields = append(fields, zap.String("tenant", tenant))
		}

		// Log error responses
		if wrapped.statusCode >= 400 {
			logger.Error("Request failed", fields...)
		} else {
			logger.Debug("Request completed", fields...)
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}
