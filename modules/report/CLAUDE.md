# Report Module

Generate and store monitoring reports.

## Current Structure

```
modules/report/
├── domain/
│   ├── entities/
│   │   └── report.go               # Report entity
│   └── repositories/
│       └── report_repository.go    # ReportRepository interface
└── infrastructure/
    ├── http/
    │   └── module.go               # All HTTP handlers inline
    └── persistence/
        └── report_postgres.go      # PostgreSQL implementation
```

No `application/` layer exists. All HTTP handlers are inline in `infrastructure/http/module.go`.

## Domain Entities

- `Report` — report with flexible content (JSON), PDF URL

## HTTP Routes (`/reports/*`, tenant-aware)

- POST `/reports` — create report (inline handler)
- GET `/reports` — list reports, optionally filtered by page (inline handler)
- GET `/reports/{id}` — get report details (inline handler)

## Infrastructure

- PostgreSQL: `reports` table (tenant-scoped) with JSON content field

## Notes

- All HTTP handlers are implemented inline in `infrastructure/http/module.go`
- No `application/` layer — no use case directories exist
- Early-stage module with minimal structure

## Architecture Improvements

- **Extract inline handlers into use cases.** Create `create_report/`, `list_reports/`, `get_report/` use case directories with handler, request, and response files.
- **Add domain repository interface.** `ReportRepository` interface exists but its usage in the HTTP layer bypasses the domain boundary.
- **Consider report generation pipeline.** Reports should be generated asynchronously (like insights) — return 202, generate in background, notify via SSE when ready.
