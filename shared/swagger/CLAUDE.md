# Swagger Package (`shared/swagger/`)

Swagger UI setup for the Chi router.

## Files

- `chi.go` — Swagger UI route configuration

## Exported API

### Functions
- `SetupSwaggerForChi(router chi.Router)` — Registers Swagger UI routes

## Routes

| Method | Path | Description |
|--------|------|-------------|
| GET | `/swagger/doc.json` | Serves swagger.json (tries `docs/swagger.json`, falls back to `/app/docs/swagger.json` for Docker) |
| GET | `/swagger` | Redirects to `/swagger/index.html` (301) |
| GET | `/swagger/*` | Serves Swagger UI assets with config: URL `/api/v1/swagger/doc.json`, deep linking, doc expansion "none" |

## Dependencies

- `chi/v5`
- `swaggo/http-swagger`

## Notes

- The `doc.json` path fallback to `/app/docs/` supports Docker container environments where the working directory differs from development
