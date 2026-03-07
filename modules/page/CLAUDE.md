# Page Module

Monitored page management (URL, name, tags).

## Domain Entities

- `Page` — page entity with workspace, URL, metadata, related check/tag data

## Use Cases (application/ directories)

- `create_page` — create a page to monitor
- `list_pages` — list pages in workspace
- `get_page` — get page details
- `update_page` — update page info
- `delete_page` — soft delete page
- `bulk_delete_pages` — delete multiple pages
- `preview_page` — page preview logic (directory exists but unused by HTTP layer)

## HTTP Routes (`/pages/*`, tenant-aware)

- POST `/pages/preview` — SSE stream proxy to extractor for live page preview (inline handler)
- POST `/pages` — create page
- POST `/pages/bulk-delete` — bulk delete pages
- GET `/pages` — list pages
- GET `/pages/{id}` — get page details
- PUT `/pages/{id}` — update page
- DELETE `/pages/{id}` — delete page

## Infrastructure

- PostgreSQL: `pages`, `page_tags` tables (tenant-scoped)
- Extractor client: proxies SSE stream from infra/extractor service for preview

## Notes

- `preview_page/` use case directory exists but the HTTP handler proxies directly to the extractor service (inline in module.go)
- Preview route uses SSE streaming (Server-Sent Events)

## Architecture Improvements

### Complete or Remove `preview_page` Use Case
The `preview_page/` directory exists but the preview handler proxies directly to the extractor service inline in `module.go`. Either move the SSE proxy logic into the use case directory or remove the empty directory.

### URL Validation
Add URL validation before storing pages:
- Validate URL format and protocol (http/https only)
- Consider DNS resolution check to catch typos early
- Block private/internal URLs (defense-in-depth alongside scraper SSRF protection)
