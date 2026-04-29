# Railway → VPS (Dokploy) Migration Design

**Date:** 2026-04-29  
**Status:** Approved

## Context

Pulzifi runs on Railway (4 services: api, worker, scraper, frontend). Migrating to a self-hosted VPS (6 vCPU / 12 GB RAM / 100 GB NVMe) using Dokploy as the orchestrator. Goal: full control, lower cost, clean dev/prod separation with no Railway artifacts.

## Target Infrastructure

### VPS Specs
- 6 vCPU, 12 GB RAM, 100 GB NVMe, 300 Mbit/s
- Dokploy installed on the VPS (manages Traefik, SSL, deployments)

### Dokploy Services (built-in, managed by Dokploy)
| Service | Version | Notes |
|---------|---------|-------|
| PostgreSQL | 17 | Persistent volume, internal only |
| Redis | 7 | Persistent volume, internal only |

### Dokploy Applications (deployed from Git via Dockerfile)
| App | Dockerfile | Public Domain | Notes |
|-----|-----------|--------------|-------|
| `api` | `Dockerfile.api` | `api.domain.com` | HTTP :3000, health `/health` |
| `worker` | `Dockerfile.worker` | none (internal) | Background jobs, `ENABLE_WORKERS=true` |
| `scraper` | `modules/infra/scraper/Dockerfile` | none (internal) | Playwright, `shm_size: 1gb` |
| `frontend` | `frontend/Dockerfile` | `domain.com` | Next.js SSR |
| `minio` | `minio/minio` official image | `storage.domain.com` | Object storage, persistent volume |

All apps communicate over Dokploy's internal Docker network. Traefik handles SSL via Let's Encrypt automatically.

## Dev / Prod Separation

### Dev Environment (local)
- Single `docker-compose.yml` (renamed from `docker-compose.monolith.yml`)
- Includes: postgres, redis, localstack (S3 emulator), scraper, monolith, worker
- All Docker service overrides (e.g. `DB_HOST: postgres`) live inline in the compose — no separate `.env.docker`
- LocalStack replaces MinIO in dev (same S3 API)
- No Railway references, no toggle comments

### Prod Environment (Dokploy)
- No compose file in the repo for prod
- Each app configured independently in Dokploy UI with env vars
- Real MinIO replaces LocalStack (`OBJECT_STORAGE_PROVIDER=minio`)
- `ENABLE_WORKERS=false` on api, `ENABLE_WORKERS=true` on worker

## CI/CD Flow

GitHub Actions (`.github/workflows/deploy.yml`):

```
push to main
  ├── go test ./...
  ├── bun run type-check (frontend)
  └── on success → trigger Dokploy deploy webhooks
        ├── api
        ├── worker
        ├── scraper
        └── frontend
```

- Webhooks stored as GitHub Secrets (`DOKPLOY_WEBHOOK_API`, `DOKPLOY_WEBHOOK_WORKER`, etc.)
- Uses GitHub Actions free tier (2000 min/month, ~130 deploys/month for this build size)
- Worker and scraper webhooks only trigger if their relevant paths changed (path filters to save minutes)

## Files Changed

| Action | File |
|--------|------|
| DELETE | `railway/` directory (4 `railway.json` files) |
| RENAME | `docker-compose.monolith.yml` → `docker-compose.yml` |
| DELETE | `.env.docker` |
| UPDATE | `docker-compose.yml` — remove Railway toggle comments, inline docker overrides, remove localstack toggle |
| UPDATE | `.env.example` — remove Railway vars, document VPS/Dokploy vars |
| CREATE | `.github/workflows/deploy.yml` |

## Environment Variables (Prod Additions)

Variables to configure in Dokploy UI (not committed):
- `MINIO_ENDPOINT` — internal service URL (e.g. `minio:9000`)
- `MINIO_PUBLIC_URL` — public URL (e.g. `https://storage.domain.com`)
- `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`
- `MINIO_USE_SSL=false` (internal traffic), `true` for public URL
- `DB_HOST` — Dokploy internal postgres hostname
- `REDIS_HOST` — Dokploy internal redis hostname
- All existing vars (JWT_SECRET, OPENROUTER_API_KEY, etc.)

## Database Migration (Railway → VPS)

One-time manual step before cutting traffic over:

```bash
# 1. Export from Railway PostgreSQL
pg_dump -h <railway-host> -U <user> -d <db> -Fc -f pulzifi-backup.dump

# 2. Copy dump to VPS
scp pulzifi-backup.dump user@vps-ip:/tmp/

# 3. Import into Dokploy PostgreSQL container
docker exec -i <postgres-container> pg_restore -U <user> -d <db> /tmp/pulzifi-backup.dump
```

Migrations run automatically on startup via `cmd/migrate/` — no manual schema setup needed.

## MinIO SSL Clarification

Internal apps (api, worker, scraper) connect via Docker network:
- `MINIO_ENDPOINT=minio:9000`
- `MINIO_USE_SSL=false`

Public clients access files via Traefik reverse proxy:
- `MINIO_PUBLIC_URL=https://storage.domain.com`
- Traefik terminates SSL; MinIO itself runs plain HTTP internally

## Dokploy Health Checks

Configure in Dokploy UI per app:

| App | Health Path | Timeout |
|-----|------------|---------|
| api | `/health` | 60s |
| scraper | `/health` | 30s |
| frontend | `/` | 60s |
| worker | none (background process) | — |
| minio | `/minio/health/live` | 30s |

## CI/CD Build Time Estimate

Per push to `main`:
- `go test ./...`: ~2-3 min
- `bun run type-check`: ~1-2 min
- Dokploy webhook triggers: ~1 min (async, Dokploy builds independently)

Total Actions minutes per deploy: ~4-6 min. At 2000 free min/month: ~330-500 deploys/month. Well within budget.

## Out of Scope

- MinIO bucket policy / CORS config (manual one-time setup via MinIO console)
- DNS setup (manual — point domain A record to VPS IP)
- Dokploy initial installation on VPS (`curl -sSL https://dokploy.com/install.sh | sh`)
