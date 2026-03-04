# Report Module

Generate and store monitoring reports.

## Domain Entities

- `Report` — report with flexible content (JSON), PDF URL

## Use Cases

- `create_report` — create a report for a page
- `list_reports` — list reports (optionally filtered by page)
- `get_report` — get report details

## HTTP Routes (`/reports/*`, tenant-aware)

- POST `/reports`
- GET `/reports`
- GET `/reports/{id}`

## Infrastructure

- PostgreSQL: `reports` table (tenant-scoped) with JSON content field
