# API Docs Module

Centralized Swagger/OpenAPI documentation hub.

## Implementation

- Uses Gin HTTP server (not Chi router)
- Aggregates Swagger docs from all microservices
- Serves documentation portal on :9000
- Proxies and rewrites service doc URLs

## Routes

- GET `/` — documentation hub with service cards
- GET `/swagger/{service}` — fetch and rewrite service docs
- GET `/swagger-initializer/{service}` — fetch initializer script

## Notes

- No database integration
- Runs independently (not part of main monolith)
