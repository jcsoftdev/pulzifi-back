# Config Package (`shared/config/`)

Environment variable loading with defaults for all application configuration.

## Files

- `config.go` — Config struct definition and `Load()` function

## Exported API

### Types
- `Config` — 43+ fields across configuration groups

### Functions
- `Load() *Config` — Loads `.env` via godotenv, reads all env vars with defaults, fatals on missing required vars

### Methods (`*Config`)
- `String() string` — Returns summary: `Config{Module: ..., DBHost: ..., HTTPPort: ..., GRPCPort: ...}`

## Configuration Groups

| Group | Fields | Required | Defaults |
|-------|--------|----------|----------|
| **Database** | `DBHost`, `DBPort`, `DBName`, `DBUser`, `DBPassword`, `DBMaxConnections` | Yes (first 5) | `DBMaxConnections`: 25 |
| **Redis** | `RedisHost`, `RedisPort`, `RedisPassword` | No | Port: 6379 |
| **Server** | `HTTPPort`, `GRPCPort`, `Environment`, `LogLevel`, `JWTSecret`, `JWTExpiration`, `JWTRefreshExpiration`, `CookieDomain` | `JWTSecret` in prod | HTTP: 3000, gRPC: 9000, JWT: 15m/7d |
| **Frontend** | `FrontendURL`, `NextJSURL`, `StaticDir` | No | NextJS: localhost:3001 |
| **CORS** | `CORSAllowedOrigins`, `CORSAllowedMethods`, `CORSAllowedHeaders` | Origins: Yes | Methods: GET,POST,PUT,DELETE,OPTIONS,PATCH |
| **MinIO/S3** | `MinIOEndpoint`, `MinIOAccessKey`, `MinIOSecretKey`, `MinIOBucket`, `MinIOUseSSL`, `MinIOPublicURL` | No | — |
| **Cloudinary** | `ObjectStorageProvider`, `CloudinaryCloudName`, `CloudinaryAPIKey`, `CloudinaryAPISecret`, `CloudinaryFolder` | No | Provider: minio |
| **Extractor** | `ExtractorURL` | Yes | — |
| **AI** | `OpenRouterAPIKey`, `OpenRouterModel`, `OpenRouterVisionModel`, `PixelDiffThreshold` | No | Model: mistralai/mistral-7b-instruct:free, Threshold: 0.001 |
| **Email** | `ResendAPIKey`, `EmailFromAddress`, `EmailFromName` | No | — |
| **OAuth** | `GoogleClientID/Secret`, `GitHubClientID/Secret`, `OAuthRedirectBaseURL` | No | — |
| **Rate Limiting** | `RateLimitRequests`, `RateLimitWindow` | No | 500 req / 60s |

## Required Environment Variables

- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`
- `CORS_ALLOWED_ORIGINS`
- `EXTRACTOR_URL`
- `JWT_SECRET` (required in production, defaults to "secret" in development with warning)

## Notes

- Supports `PORT` env var as fallback for `HTTP_PORT` (Railway compatibility)
- Supports `DATABASE_URL` env var for direct connection string (Railway)
- Uses `godotenv` to load `.env` file (non-fatal if missing)
