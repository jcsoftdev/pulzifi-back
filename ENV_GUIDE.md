# Environment Variables by Docker Service

This document maps which environment variables are used by each service/Docker container to help you configure different stages for deployment without mixing all variables together.

---

## 1. **API Server** (`Dockerfile.api`)

**Production build of `cmd/server/main.go` HTTP server**

### Database
- `DB_HOST` [REQUIRED]
- `DB_PORT` [REQUIRED]
- `DB_NAME` [REQUIRED]
- `DB_USER` [REQUIRED]
- `DB_PASSWORD` [REQUIRED]
- `DB_MAX_CONNECTIONS` (default: 25)

### Server & Auth
- `HTTP_PORT` (default: 3000)
- `GRPC_PORT` (default: 9000)
- `ENVIRONMENT` (development/production)
- `LOG_LEVEL` (debug/info/warn/error)
- `JWT_SECRET` [REQUIRED in production]
- `JWT_EXPIRATION` (default: 3600s / 1 hour)
- `JWT_REFRESH_EXPIRATION` (default: 604800s / 7 days)

### Frontend & Routing
- `FRONTEND_URL` [REQUIRED] — e.g., https://app.pulzifi.com
- `NEXTJS_URL` (default: http://localhost:3001)
- `COOKIE_DOMAIN` — e.g., .pulzifi.com (for cross-subdomain cookies)
- `CORS_ALLOWED_ORIGINS` [REQUIRED] — comma-separated list

### Scraper/Extractor
- `EXTRACTOR_URL` [REQUIRED] — HTTP endpoint of scraper service
  - Docker: `http://scraper:3000`
  - Production: depends on scraper deployment

### Object Storage
- `OBJECT_STORAGE_PROVIDER` (default: minio) — `minio` or `cloudinary`
- **MinIO/S3 variants:**
  - `MINIO_ENDPOINT` — e.g., minio:9000 (Docker) or s3-endpoint (production)
  - `MINIO_ACCESS_KEY`
  - `MINIO_SECRET_KEY`
  - `MINIO_BUCKET` (default: pulzifi-snapshots)
  - `MINIO_USE_SSL` (default: false)
  - `MINIO_PUBLIC_URL` — e.g., http://localhost:9000
- **Cloudinary variants (if PROVIDER=cloudinary):**
  - `CLOUDINARY_CLOUD_NAME`
  - `CLOUDINARY_API_KEY`
  - `CLOUDINARY_API_SECRET`
  - `CLOUDINARY_FOLDER` (default: pulzifi)

### AI Insights (Optional)
- `OPENROUTER_API_KEY` — LLM service API key
- `OPENROUTER_MODEL` (default: mistralai/mistral-7b-instruct:free)
- `OPENROUTER_VISION_MODEL` — for image analysis
- `PIXEL_DIFF_THRESHOLD` (default: 0.001)

### Email Notifications (Optional)
- `RESEND_API_KEY` — Resend email service API key
- `EMAIL_FROM_ADDRESS` — e.g., noreply@pulzifi.com
- `EMAIL_FROM_NAME` — e.g., Pulzifi

### OAuth (Optional)
- `GOOGLE_CLIENT_ID`
- `GOOGLE_CLIENT_SECRET`
- `GITHUB_CLIENT_ID`
- `GITHUB_CLIENT_SECRET`
- `OAUTH_REDIRECT_BASE_URL` — e.g., https://app.pulzifi.com

### Redis (Optional)
- `REDIS_HOST` — IP/hostname of Redis server
- `REDIS_PORT` (default: 6379)
- `REDIS_PASSWORD` — if Redis requires authentication

---

## 2. **Worker** (`Dockerfile.worker`)

**Production build of `cmd/worker/main.go` background job processor**

Uses **the same environment variables as the API Server**, but only reads:

### Always Used
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`, `DB_MAX_CONNECTIONS`
- `ENVIRONMENT`, `LOG_LEVEL`
- `EXTRACTOR_URL` — to call scraper for snapshot capture
- `OBJECT_STORAGE_PROVIDER` + MinIO/Cloudinary credentials (for snapshot storage)

### Conditionally Used
- `OPENROUTER_API_KEY`, `OPENROUTER_MODEL`, `OPENROUTER_VISION_MODEL`, `PIXEL_DIFF_THRESHOLD` — for AI insight generation
- `RESEND_API_KEY`, `EMAIL_FROM_ADDRESS`, `EMAIL_FROM_NAME` — for sending alert emails
- `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD` — for deduplication and locking

### NOT Used by Worker
- `HTTP_PORT`, `GRPC_PORT` — worker doesn't listen for HTTP
- `JWT_SECRET`, `JWT_EXPIRATION`, etc. — worker doesn't handle auth
- `FRONTEND_URL`, `NEXTJS_URL`, `COOKIE_DOMAIN`, `CORS_ALLOWED_ORIGINS` — not applicable to worker
- `GOOGLE_CLIENT_ID/SECRET`, `GITHUB_CLIENT_ID/SECRET`, `OAUTH_REDIRECT_BASE_URL` — worker doesn't handle OAuth

---

## 3. **Scraper Service** (`modules/infra/scraper/Dockerfile`)

**TypeScript/Bun Playwright-based web scraper (standalone service)**

### Port Configuration
- `PORT` (default: 3000) — HTTP port for extraction/preview endpoints

### Browser & Performance
- `MAX_CONCURRENT_PAGES` (default: 3) — max concurrent page extractions
- `NAV_TIMEOUT_MS` (default: 30000) — milliseconds to wait for page navigation
- `CF_MAX_WAIT_MS` (default: 20000) — milliseconds to wait for Cloudflare challenge
- `SCREENSHOT_QUALITY` (default: 80) — JPEG quality (1-100)

### Chromium
- `CHROMIUM_PATH` (default: /usr/bin/chromium) — path to system Chromium binary
- `SHM_SIZE` (Docker compose: 1gb) — shared memory for Chromium (set in docker-compose)

### NOT Used by Scraper
- **No database, Redis, authentication, email, or storage configuration needed**
- Scraper is stateless and only handles HTTP requests for page extraction/preview

---

## 4. **Frontend** (`frontend/Dockerfile`)

**Next.js 16 application (Bun runtime)**

### Build-Time Variables (baked into bundle)
- `NEXT_PUBLIC_APP_DOMAIN` — e.g., pulzifi.com (used for subdomain validation)
- `NEXT_PUBLIC_APP_BASE_URL` — e.g., https://app.pulzifi.com (for cross-subdomain redirects)
- `SERVER_API_URL` — e.g., https://api.pulzifi.com (backend API base URL, if different from frontend domain)

### Runtime Configuration
- `PORT` (default: 3000) — HTTP port for Next.js server
- `NODE_ENV` (set to production in Dockerfile)

### NOT Used by Frontend
- Database, Redis, Extractor, Storage, AI, Email, OAuth credentials — all handled by backend API

---

## 5. **PostgreSQL** (`docker-compose.monolith.yml`)

**Postgres 17 service (development only)**

### Database Initialization
- `POSTGRES_DB` (default: pulzifi) — initial database name
- `POSTGRES_USER` (default: pulzifi_user) — superuser name
- `POSTGRES_PASSWORD` (default: pulzifi_password) — superuser password
- `POSTGRES_INITDB_ARGS` — encoding and locale settings

### Notes
- **Not deployed to production** — Railway uses managed PostgreSQL
- Postgres reads these from Docker environment in compose file, not from .env
- In production, API/Worker servers connect to Railway's managed database via `DB_HOST` etc.

---

## 6. **LocalStack** (MinIO S3 emulation, `docker-compose.monolith.yml`)

**AWS S3 emulator for local development (development only)**

### LocalStack Configuration
- `SERVICES=s3` — only S3 service enabled
- `DEBUG=1` — verbose logging

### Used By
- API Server + Worker read snapshot bucket from here if `OBJECT_STORAGE_PROVIDER=minio`
- Endpoint: `MINIO_ENDPOINT=localstack:4566`
- Credentials: `MINIO_ACCESS_KEY=test`, `MINIO_SECRET_KEY=test` (hardcoded in compose)

### Notes
- **Not deployed to production** — production uses real MinIO or Cloudinary
- LocalStack is purely for local development

---

## 7. **Redis** (`docker-compose.monolith.yml`)

**Redis 7 service (currently commented out in compose)**

### Configuration
- `REDIS_HOST` — service name or IP
- `REDIS_PORT` (default: 6379)
- `REDIS_PASSWORD` — if authentication enabled

### Used By
- API Server + Worker for:
  - **Refresh token deduplication** — 2-second grace period to prevent duplicate tokens
  - **Rate limiting** — per-IP token bucket (currently in-memory, needs Redis for multi-instance)
  - **Nonce store** — temporary storage for cross-subdomain auth flow
  - **Session caching** — optional user session caching
  - **Pub/Sub brokers cache** — SSE broker cache replication across instances

### Notes
- **Optional in development** — application works without Redis but with limited features
- **Should be required in production** — for multi-instance deployments
- Currently commented out in docker-compose.monolith.yml (uncomment if needed)

---

## Stage-Specific Configuration

### Development (localhost)
**Single `docker-compose.monolith.yml` with all services**
- Use `.env` + `.env.docker` (overrides)
- DB: local postgres service
- Storage: localstack (S3 emulation)
- Scraper: local container
- Redis: optional (commented out)
- All credentials can be dummy values

### Staging (Cloud VMs)
**Separate containers, managed network**
```
API Server (Dockerfile.api)
├─ DB_HOST → staging-db.internal:5432
├─ EXTRACTOR_URL → http://scraper-vm:3000
├─ MINIO_ENDPOINT → staging-s3.internal:9000
└─ REDIS_HOST → staging-redis.internal:6379

Worker (Dockerfile.worker)
├─ Same DB + Storage + Scraper URLs

Scraper (modules/infra/scraper/Dockerfile)
├─ PORT → 3000

Frontend (frontend/Dockerfile)
├─ NEXT_PUBLIC_APP_DOMAIN → staging.pulzifi.com
└─ SERVER_API_URL → https://api.staging.pulzifi.com
```

### Production (Railway or Kubernetes)
**Managed services**
```
API Server (Dockerfile.api)
├─ DB_HOST → shinkansen.proxy.rlwy.net (Railway managed PostgreSQL)
├─ EXTRACTOR_URL → https://scraper.production.internal
├─ MINIO_ENDPOINT → production-s3.internal:9000 (or S3/Cloudinary)
├─ REDIS_HOST → production-redis.internal:6379 (required)
└─ JWT_SECRET, CORS_ALLOWED_ORIGINS, etc. → production values

Worker (Dockerfile.worker)
├─ Same as API

Scraper (modules/infra/scraper/Dockerfile)
├─ PORT → 3000 (internal)
├─ MAX_CONCURRENT_PAGES → 5+ (increase for throughput)

Frontend (frontend/Dockerfile)
├─ NEXT_PUBLIC_APP_DOMAIN → pulzifi.com
└─ SERVER_API_URL → https://api.pulzifi.com
```

---

## Summary Table

| Service | Docker | DB | Storage | Scraper | Redis | Email | OAuth | AI |
|---------|--------|----|---------|---------|----- |-------|-------|-----|
| **API** | ✅ | ✅ | ✅ | ✅ | ⭕ | ✅ | ✅ | ✅ |
| **Worker** | ✅ | ✅ | ✅ | ✅ | ⭕ | ✅ | ❌ | ✅ |
| **Scraper** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Frontend** | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Postgres** | ✅ | N/A | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **LocalStack** | ✅ | ❌ | N/A | ❌ | ❌ | ❌ | ❌ | ❌ |
| **Redis** | ✅ | ❌ | ❌ | ❌ | N/A | ❌ | ❌ | ❌ |

✅ = Required | ⭕ = Optional | ❌ = Not used | N/A = Built-in function

---

## Configuration Files

- **`.env.example`** — Template with all possible variables and descriptions
- **`.env.docker`** — Docker compose overrides (DB_HOST: postgres/Railway, EXTRACTOR_URL: http://scraper:3000)
- **`.env`** — Your actual values (git-ignored) — merge of above with your secrets

For staging/production, create separate config files per environment and pass via `--env-file` flag or deployment platform's secret manager (Railway, Kubernetes, etc.).

---

## Quick Reference: Which Service Needs What

### Just deploying API + Frontend?
- API: DB, EXTRACTOR_URL, MINIO/Cloudinary, JWT_SECRET, CORS_ALLOWED_ORIGINS
- Frontend: NEXT_PUBLIC_APP_DOMAIN, NEXT_PUBLIC_APP_BASE_URL, SERVER_API_URL

### Adding background jobs (Worker)?
- Worker: Same as API (DB, EXTRACTOR_URL, Storage)
- Consider enabling Redis for deduplication and rate limiting

### Multi-instance deployment?
- **Enable Redis** for API + Worker:
  - Rate limiting shared across instances
  - Refresh token deduplication
  - Nonce store persistence
  - SSE broker cache replication
- **Scale Scraper** with load balancer in front of multiple instances

### Disable/Remove unused features:
- **No OAuth?** Skip GOOGLE_CLIENT_ID/SECRET, GITHUB_CLIENT_ID/SECRET, OAUTH_REDIRECT_BASE_URL
- **No AI Insights?** Skip OPENROUTER_API_KEY, OPENROUTER_MODEL, OPENROUTER_VISION_MODEL, PIXEL_DIFF_THRESHOLD
- **No Email Alerts?** Skip RESEND_API_KEY, EMAIL_FROM_ADDRESS, EMAIL_FROM_NAME
