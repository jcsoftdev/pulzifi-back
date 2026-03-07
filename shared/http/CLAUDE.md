# HTTP Package (`shared/http/`)

Response helpers and Chi middleware utilities.

## Files

- `response.go` — JSON/XML/HTML/Text response helpers with status shortcuts
- `middleware.go` — Chi middleware (logging, recovery, request ID, timeout, CORS)

## Exported API

### Response Helper (`response.go`)

**Structs:**
- `ResponseHelper` — Wraps `*zap.Logger` for response methods

**Functions:**
- `NewResponseHelper(logger) *ResponseHelper` — Creates helper with logger
- `SetDefaultLogger(logger)` — Sets logger for package-level convenience functions
- `RespondJSON(w, statusCode, data)` — Package-level JSON response
- `RespondError(w, statusCode, message)` — Package-level error response (`{"error": message}`)

**Methods (`*ResponseHelper`):**
- `RespondJSON(w, statusCode, data)` — JSON response
- `RespondError(w, statusCode, message)` — Error JSON
- `RespondXML(w, statusCode, data)` — XML response
- `RespondHTML(w, statusCode, templateStr, data)` — HTML template response
- `RespondText(w, statusCode, text)` — Plain text response
- `RespondNoContent(w)` — 204
- `RespondCreated(w, data)` — 201
- `RespondOK(w, data)` — 200
- `RespondBadRequest(w, message)` — 400
- `RespondUnauthorized(w, message)` — 401
- `RespondForbidden(w, message)` — 403
- `RespondNotFound(w, message)` — 404
- `RespondConflict(w, message)` — 409
- `RespondInternalServerError(w, message)` — 500
- `RespondNotImplemented(w, message)` — 501

### Chi Middleware (`middleware.go`)

**Structs:**
- `ChiMiddleware` — Wraps `*zap.Logger`

**Functions:**
- `NewChiMiddleware(logger) *ChiMiddleware`

**Methods (`*ChiMiddleware`):**
- `Logging()` — HTTP request logging via `httplog.RequestLogger`
- `Recovery()` — Panic recovery, logs with zap, returns 500
- `RequestID()` — Reads `X-Request-ID` header or generates timestamp-based ID
- `Timeout(duration)` — Wraps with `http.TimeoutHandler`
- `CORS(allowedOrigins)` — Sets CORS headers, handles OPTIONS preflight
- `ContentType(contentType)` — Sets Content-Type header

## Dependencies

- `go-chi/httplog/v2`, `zap`
