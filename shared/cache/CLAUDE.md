# Cache Package (`shared/cache/`)

Redis client singleton and refresh token caching.

## Files

- `redis.go` — Redis client initialization and lifecycle
- `refresh_token_cache.go` — Concurrent refresh token deduplication cache

## Exported API

### Redis Client (`redis.go`)
- `InitRedis(cfg *config.Config) error` — Initializes Redis client with 5s timeout ping test
- `GetRedisClient() *redis.Client` — Returns the singleton client
- `CloseRedis() error` — Closes connection (nil-safe)

### Refresh Token Cache (`refresh_token_cache.go`)
- `RefreshTokenGracePeriod = 2 * time.Second` — Cache TTL for deduplication
- `RefreshTokenCache` — Cached response struct: `AccessToken`, `RefreshToken`, `ExpiresIn`, `Tenant`
- `GetRefreshTokenCache(ctx, oldRefreshToken) (*RefreshTokenCache, error)` — Retrieve cached response
- `SetRefreshTokenCache(ctx, oldRefreshToken, cache) error` — Store with 2s TTL
- `DeleteRefreshTokenCache(ctx, oldRefreshToken) error` — Remove entry

## Key Pattern

`refresh_token_cache:<oldRefreshToken>`

## Graceful Degradation

All refresh token cache functions silently return `nil`/no-op if Redis client is not initialized. The application works without Redis (concurrent refresh requests may produce duplicate tokens but won't fail).

## Dependencies

- `go-redis/v9`
- `shared/config`

## Architecture Improvements

### Production Redis Requirement
Redis is currently optional with graceful degradation. For production, Redis should be **required** because:
- Without Redis, concurrent refresh token requests can produce duplicate tokens (race condition)
- Rate limiting state (`shared/middleware/rate_limiter.go`) is per-instance without Redis, allowing clients to bypass limits by hitting different instances
- Session caching and nonce storage are node-local without a shared backend

### Additional Redis-Backed Features to Add
- **Rate limiter state sharing:** Move `shared/middleware/rate_limiter.go` token buckets to Redis (`INCR`/`EXPIRE` or Redis Cell module)
- **Nonce store:** Replace `shared/noncestore/` in-memory map with Redis `SETEX` (30s TTL)
- **SSE broker cache:** Replace `shared/pubsub/` in-memory caches with Redis `SETEX`
- **Session store:** Cache user sessions to reduce database lookups on every authenticated request

### Connection Pool Sizing
Currently uses default `go-redis` pool settings. For high-traffic production, configure:
- `PoolSize` based on expected concurrent connections
- `MinIdleConns` for warm connection pool
- `DialTimeout`, `ReadTimeout`, `WriteTimeout` for predictable latency
