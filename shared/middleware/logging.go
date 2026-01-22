package middleware

import (
	"net/http"
	"time"

	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// LoggingMiddleware logs all incoming requests with timing information
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		path := r.URL.Path
		method := r.Method
		tenant := GetTenantFromContext(r.Context())
		hostname := r.Header.Get("host")

		if tenant == "" {
			logger.Warn("Request received with empty tenant",
				zap.String("method", method),
				zap.String("path", path),
				zap.String("hostname", hostname),
				zap.String("remote_addr", r.RemoteAddr))
		} else {
			logger.Debug("Request started",
				zap.String("method", method),
				zap.String("path", path),
				zap.String("tenant", tenant),
				zap.String("hostname", hostname),
				zap.String("remote_addr", r.RemoteAddr))
		}

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriterWrapper{ResponseWriter: w}

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(startTime)

		if wrapped.statusCode >= 400 {
			logger.Error("Request failed",
				zap.String("method", method),
				zap.String("path", path),
				zap.String("tenant", tenant),
				zap.Int("status", wrapped.statusCode),
				zap.Duration("duration", duration))
		} else {
			logger.Info("Request completed",
				zap.String("method", method),
				zap.String("path", path),
				zap.String("tenant", tenant),
				zap.Int("status", wrapped.statusCode),
				zap.Duration("duration", duration))
		}
	})
}

// responseWriterWrapper wraps http.ResponseWriter to capture status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}
