# Report Module

Generate and store monitoring reports.

## Domain Entities

- `Report` — report with flexible content (JSON), PDF URL

## Use Cases (application/ directories)

- `create_report` — create a report (directory exists but unused by HTTP layer)

## HTTP Routes (`/reports/*`, tenant-aware)

- POST `/reports` — create report (inline handler)
- GET `/reports` — list reports, optionally filtered by page (inline handler)
- GET `/reports/{id}` — get report details (inline handler)

## Infrastructure

- PostgreSQL: `reports` table (tenant-scoped) with JSON content field

## Notes

- All HTTP handlers are implemented inline in module.go
- The `create_report/` use case directory exists but is not referenced by the HTTP layer
- Early-stage module with minimal structure

## Architecture Improvements

- **Extract inline handlers into use cases.** Create `create_report/`, `list_reports/`, `get_report/` use case directories with handler, request, and response files.
- **Add domain entities.** The `Report` entity exists but the module lacks proper domain modeling. Define repository interfaces and move SQL out of module.go.
- **Consider report generation pipeline.** Reports should be generated asynchronously (like insights) — return 202, generate in background, notify via SSE when ready.
