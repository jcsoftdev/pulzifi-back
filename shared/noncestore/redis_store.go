package noncestore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	redisNonceTTL    = 30 * time.Second
	redisNoncePrefix = "nonce:"
)

// RedisStore is a Redis-backed nonce store for multi-instance deployments.
// It implements NonceStore and is safe for concurrent use from multiple goroutines
// and across multiple server instances.
//
// Key format: nonce:<id>  →  JSON-encoded NonceEntry  EX 30
// Consume uses GETDEL (Redis 6.2+) for an atomic get-and-delete.
type RedisStore struct {
	rdb *redis.Client
}

// NewRedisStore creates a Redis-backed NonceStore using the provided client.
// The client is typically obtained from shared/cache.GetRedisClient().
func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{rdb: rdb}
}

// Save stores a nonce entry in Redis with a 30-second TTL.
// Storing the same ID a second time overwrites the previous entry and resets the TTL.
func (s *RedisStore) Save(nonce string, entry NonceEntry) {
	ctx := context.Background()
	entry.CreatedAt = time.Now()

	data, err := json.Marshal(entry)
	if err != nil {
		return // malformed entry — silently discard
	}

	s.rdb.Set(ctx, redisNoncePrefix+nonce, data, redisNonceTTL)
}

// Consume atomically retrieves and deletes the nonce entry (GETDEL — single round-trip).
// Returns nil if the nonce does not exist or has expired.
func (s *RedisStore) Consume(nonce string) *NonceEntry {
	ctx := context.Background()

	data, err := s.rdb.GetDel(ctx, redisNoncePrefix+nonce).Bytes()
	if err != nil {
		// redis.Nil means key does not exist (expired or never saved).
		return nil
	}

	var entry NonceEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil
	}

	return &entry
}

// Peek reads the nonce entry without consuming it.
// Returns nil if the nonce does not exist or has expired (Redis handles TTL automatically).
func (s *RedisStore) Peek(nonce string) *NonceEntry {
	ctx := context.Background()

	data, err := s.rdb.Get(ctx, redisNoncePrefix+nonce).Bytes()
	if err != nil {
		return nil
	}

	var entry NonceEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil
	}

	return &entry
}
