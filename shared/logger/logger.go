package logger

import (
	"context"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

type contextKey string

const CorrelationIDKey contextKey = "correlation_id"
const TenantKey contextKey = "tenant"
const UserIDKey contextKey = "user_id"

func init() {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(parseLevel(level))
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true // Disable stack trace by default for cleaner logs

	var err error
	Logger, err = config.Build()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// extractContextFields extracts logging context from context.Context
func extractContextFields(ctx context.Context) []zap.Field {
	var fields []zap.Field

	if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok && correlationID != "" {
		fields = append(fields, zap.String("correlation_id", correlationID))
	}

	if tenant, ok := ctx.Value(TenantKey).(string); ok && tenant != "" {
		fields = append(fields, zap.String("tenant", tenant))
	}

	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	return fields
}

func InfoWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := append(extractContextFields(ctx), fields...)
	Logger.Info(msg, allFields...)
}

func ErrorWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := append(extractContextFields(ctx), fields...)
	Logger.Error(msg, allFields...)
}

func DebugWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := append(extractContextFields(ctx), fields...)
	Logger.Debug(msg, allFields...)
}

func WarnWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := append(extractContextFields(ctx), fields...)
	Logger.Warn(msg, allFields...)
}

// Legacy convenience functions
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}
