# Snapshot Module

Page snapshot capture and storage (worker service).

## Domain Entities

- `SnapshotRequest` — snapshot request from queue
- `SnapshotResult` — snapshot result (image, HTML, status)

## Application Services

- `SnapshotWorker` — main orchestrator
- `SnapshotService` — capture and upload logic

## Infrastructure

- Extractor client: HTTP call to infra/extractor service
- Object storage: MinIO or Cloudinary for image/HTML storage
- Email notifications: sends change notification emails
- Insight generation: async insight generation after changes detected
- Event bus: publishes snapshot-completed events
- Webhook publishing for integrations

## Notes

- Detects changes via content hash comparison
- No HTTP routes — runs as background worker
