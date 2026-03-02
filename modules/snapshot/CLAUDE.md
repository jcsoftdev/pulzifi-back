# Snapshot Module

## Responsibility

Screenshot capture via Playwright, HTML extraction, object storage upload (Cloudinary/MinIO), and content hash computation for change detection.

## Entities

- **SnapshotRequest** — PageID, URL, SchemaName
- **SnapshotResult** — PageID, URL, ImageURL, HTMLURL, TextURL, ImageHash, HTMLHash, TextHash, ContentHash, Status, ErrorMessage

## Repository Interfaces

- `ObjectStorage` — Upload, EnsureBucket

## Infrastructure

- **Extractor client**: HTTP client to the Playwright Node.js service (`modules/infra/extractor/`)
- **Cloudinary provider**: Cloud image/file storage
- **MinIO provider**: S3-compatible local storage (LocalStack in dev)

## Dependencies

- Monitoring module (provides check context)
- Insight module (triggers AI analysis after change detection)
- Email module (notification on changes)
- EventBus (publishes snapshot events)
- External: Playwright extractor service on port 3005

## Constraints

- Extractor service must be running and healthy
- Content hash is SHA256 of normalized HTML text
- Screenshots stored as PNG in object storage
- HTML snapshots stored as compressed text
