package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ErrCacheDisabled is returned by cache functions when Redis is not initialized.
// Callers should treat this as a cache miss — not as a hard failure.
var ErrCacheDisabled = errors.New("cache: Redis is not initialized")

const (
	// Grace period for refresh token reuse (2 seconds to handle concurrent requests)
	RefreshTokenGracePeriod = 2 * time.Second
)

// RefreshTokenCache represents cached refresh token response
type RefreshTokenCache struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Tenant       string `json:"tenant,omitempty"`
}

// GetRefreshTokenCache retrieves cached refresh token response.
// Returns (nil, ErrCacheDisabled) when Redis is not initialized so callers can
// distinguish a genuine cache miss (err != nil) from a cached hit (nil error).
func GetRefreshTokenCache(ctx context.Context, oldRefreshToken string) (*RefreshTokenCache, error) {
	if redisClient == nil {
		return nil, ErrCacheDisabled
	}

	key := fmt.Sprintf("refresh_token_cache:%s", oldRefreshToken)
	data, err := redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var cache RefreshTokenCache
	if err := json.Unmarshal([]byte(data), &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

// SetRefreshTokenCache stores refresh token response in cache
func SetRefreshTokenCache(ctx context.Context, oldRefreshToken string, cache *RefreshTokenCache) error {
	if redisClient == nil {
		return nil // Redis disabled, silently ignore
	}

	key := fmt.Sprintf("refresh_token_cache:%s", oldRefreshToken)
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}

	return redisClient.Set(ctx, key, data, RefreshTokenGracePeriod).Err()
}

// DeleteRefreshTokenCache removes cached refresh token response
func DeleteRefreshTokenCache(ctx context.Context, oldRefreshToken string) error {
	if redisClient == nil {
		return nil // Redis disabled, silently ignore
	}

	key := fmt.Sprintf("refresh_token_cache:%s", oldRefreshToken)
	return redisClient.Del(ctx, key).Err()
}
