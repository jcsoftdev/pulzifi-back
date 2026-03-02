# Report Module

## Responsibility

Report generation and management from monitoring data. Reports summarize page monitoring activity over a time period.

## Entities

- **Report** — ID, PageID, Title, ReportDate, Content, PDFURL, CreatedBy, CreatedAt, DeletedAt

## Repository Interfaces

- `ReportRepository` — Create, GetByID, ListByPage, List

## Routes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/reports` | Create report |
| GET | `/reports` | List reports (filterable by page_id) |
| GET | `/reports/{id}` | Get report |

## Dependencies

- Page module
- Auth middleware

## Constraints

- Tenant-scoped
