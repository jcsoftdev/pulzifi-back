# Snapshot Module

Page snapshot capture and storage (worker service).

## Domain Entities

- `SnapshotRequest` — snapshot request with page ID, URL, and schema name
- `SnapshotResult` — snapshot result (image URL, HTML URL, text URL, content hashes, status)

## Domain Services

- `ImageCompare` — pixel-based image comparison for change detection (`image_compare.go`, `image_compare_test.go`)

## Application Services (flat files, no use case directories)

- `worker.go` — main snapshot worker orchestrator
- `snapshot_service.go` — capture and upload logic

## Infrastructure

- Extractor client: HTTP call to infra/scraper service (`infrastructure/extractor/client.go`)
- Object storage provider: `infrastructure/storage/provider.go` (selects MinIO or Cloudinary; accepts `""`, `"minio"`, `"s3"` for MinIO, `"cloudinary"` for Cloudinary)
- MinIO client: `infrastructure/minio/client.go` (S3-compatible storage)
- Cloudinary client: `infrastructure/cloudinary/client.go` (cloud image storage)
- `infrastructure/kafka/` — exists but empty (Kafka consumer placeholder)

## Notes

- Detects changes via content hash comparison + pixel diff threshold
- No HTTP routes — runs as background worker only
- Has its own `Dockerfile` for standalone deployment
- Has a `cmd/test/` directory with a test entry point
- Application layer uses flat files instead of one-use-case-per-directory pattern

## Architecture Improvements

- **Flat application layer should be structured.** Extract into use cases: `capture_snapshot/`, `upload_snapshot/`, `compare_snapshots/` to match the project convention.
- **Empty kafka/ directory** should either be implemented (for durable event consumption) or removed.
- **Storage provider selection** at startup is good, but the provider could be injected rather than selected via switch statement for better testability.
- **Image comparison** is CPU-bound. For high throughput, consider offloading to a separate worker pool or using a more efficient comparison algorithm (perceptual hashing instead of pixel diff).
