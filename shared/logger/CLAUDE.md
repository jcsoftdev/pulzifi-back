# Logger Package (`shared/logger/`)

Zap structured logging with context-aware enrichment.

## Files

- `logger.go` — Global logger initialization and context-aware logging functions

## Exported API

### Variables
- `Logger *zap.Logger` — Global logger, initialized in `init()` with production config, ISO8601 time encoding

### Context Keys
- `CorrelationIDKey` — Extracts `correlation_id` from context
- `TenantKey` — Extracts `tenant` from context
- `UserIDKey` — Extracts `user_id` from context

### Context-Aware Functions
- `InfoWithContext(ctx, msg, ...fields)` — Logs at info level with context fields
- `ErrorWithContext(ctx, msg, ...fields)` — Logs at error level with context fields
- `DebugWithContext(ctx, msg, ...fields)` — Logs at debug level with context fields
- `WarnWithContext(ctx, msg, ...fields)` — Logs at warn level with context fields

### Direct Functions (Legacy)
- `Info(msg, ...fields)`
- `Error(msg, ...fields)`
- `Debug(msg, ...fields)`
- `Warn(msg, ...fields)`

## Configuration

Log level set from `LOG_LEVEL` env var. Supported values: `debug`, `info`, `warn`, `error` (default: `info`).

## Notes

- Context-aware functions automatically extract `correlation_id`, `tenant`, and `user_id` from request context
- Uses Zap production config with stack traces disabled by default
- Prefer `*WithContext` variants over direct functions for request-scoped logging
