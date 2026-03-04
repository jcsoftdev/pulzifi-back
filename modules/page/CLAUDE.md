# Page Module

Monitored page management (URL, name, tags).

## Domain Entities

- `Page` — page entity with workspace, URL, metadata, related check/tag data

## Use Cases

- `create_page` — create a page to monitor
- `list_pages` — list pages in workspace
- `get_page` — get page details
- `update_page` — update page info
- `delete_page` — soft delete page
- `bulk_delete_pages` — delete multiple pages

## HTTP Routes (`/pages/*`, tenant-aware)

- POST `/pages`
- GET `/pages`
- GET `/pages/{id}`
- PUT `/pages/{id}`
- DELETE `/pages/{id}`
- POST `/pages/bulk-delete`

## Infrastructure

- PostgreSQL: `pages`, `page_tags` tables (tenant-scoped)
