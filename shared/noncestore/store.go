package noncestore

import (
	"sync"
	"time"
)

const nonceTTL = 30 * time.Second

// NonceEntry holds the JWT tokens stored under a one-time nonce.
type NonceEntry struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	CreatedAt    time.Time
}

// Store is a short-lived in-memory store for cross-subdomain cookie exchange.
// After login, tokens are stored under a nonce; the tenant callback consumes
// the nonce and sets HttpOnly cookies on the tenant origin.
type Store struct {
	mu      sync.Mutex
	entries map[string]NonceEntry
}

// New creates a new nonce store.
func New() *Store {
	return &Store{
		entries: make(map[string]NonceEntry),
	}
}

// Save stores a nonce entry and lazily prunes expired entries.
func (s *Store) Save(nonce string, entry NonceEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for k, v := range s.entries {
		if now.Sub(v.CreatedAt) > nonceTTL {
			delete(s.entries, k)
		}
	}

	entry.CreatedAt = now
	s.entries[nonce] = entry
}

// Consume retrieves and deletes a nonce entry (one-time use).
// Returns nil if the nonce does not exist or has expired.
func (s *Store) Consume(nonce string) *NonceEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.entries[nonce]
	if !ok {
		return nil
	}
	delete(s.entries, nonce)

	if time.Since(entry.CreatedAt) > nonceTTL {
		return nil
	}

	return &entry
}

// Peek reads a nonce entry without consuming it.
// Returns nil if the nonce does not exist or has expired.
func (s *Store) Peek(nonce string) *NonceEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.entries[nonce]
	if !ok {
		return nil
	}

	if time.Since(entry.CreatedAt) > nonceTTL {
		delete(s.entries, nonce)
		return nil
	}

	return &entry
}
