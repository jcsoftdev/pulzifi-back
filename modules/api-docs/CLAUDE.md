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

## Architecture Improvements

### Consolidate Into Monolith
This module uses Gin on :9000 while the rest of the app uses Chi on :3000. Consider:
- Porting the Swagger aggregation to Chi and mounting under `/docs` in the monolith
- This eliminates a separate service to deploy and the Gin dependency for this single purpose
- The monolith already has `shared/swagger/` for Chi-based Swagger UI

### Auto-Discovery
Service URLs and names are hardcoded. Consider auto-discovering registered modules via the `shared/router/` module registry.
