package noncestore

import (
	"testing"
	"time"
)

func TestSaveAndConsume(t *testing.T) {
	s := New()
	entry := NonceEntry{
		AccessToken:  "at-123",
		RefreshToken: "rt-456",
		ExpiresIn:    900,
	}
	s.Save("nonce-1", entry)

	got := s.Consume("nonce-1")
	if got == nil {
		t.Fatal("expected entry, got nil")
	}
	if got.AccessToken != "at-123" {
		t.Errorf("AccessToken = %q, want %q", got.AccessToken, "at-123")
	}
	if got.RefreshToken != "rt-456" {
		t.Errorf("RefreshToken = %q, want %q", got.RefreshToken, "rt-456")
	}
	if got.ExpiresIn != 900 {
		t.Errorf("ExpiresIn = %d, want 900", got.ExpiresIn)
	}

	// Second consume should return nil (one-time use)
	if s.Consume("nonce-1") != nil {
		t.Error("expected nil on second consume")
	}
}

func TestPeekDoesNotConsume(t *testing.T) {
	s := New()
	s.Save("nonce-2", NonceEntry{AccessToken: "at-peek"})

	got := s.Peek("nonce-2")
	if got == nil || got.AccessToken != "at-peek" {
		t.Fatal("Peek should return the entry")
	}

	// Peek again — should still be there
	got = s.Peek("nonce-2")
	if got == nil || got.AccessToken != "at-peek" {
		t.Fatal("Peek should not consume the entry")
	}

	// Consume should still work
	got = s.Consume("nonce-2")
	if got == nil || got.AccessToken != "at-peek" {
		t.Fatal("Consume should return the entry after peeks")
	}
}

func TestConsumeNonExistent(t *testing.T) {
	s := New()
	if s.Consume("does-not-exist") != nil {
		t.Error("expected nil for non-existent nonce")
	}
}

func TestPeekNonExistent(t *testing.T) {
	s := New()
	if s.Peek("does-not-exist") != nil {
		t.Error("expected nil for non-existent nonce")
	}
}

func TestExpiredEntry(t *testing.T) {
	s := New()
	entry := NonceEntry{
		AccessToken:  "expired-token",
		RefreshToken: "expired-refresh",
		ExpiresIn:    900,
		CreatedAt:    time.Now().Add(-31 * time.Second), // already expired
	}

	s.mu.Lock()
	s.entries["expired-nonce"] = entry
	s.mu.Unlock()

	if s.Consume("expired-nonce") != nil {
		t.Error("expected nil for expired nonce on Consume")
	}

	// Re-insert for Peek test
	s.mu.Lock()
	s.entries["expired-nonce-2"] = entry
	s.mu.Unlock()

	if s.Peek("expired-nonce-2") != nil {
		t.Error("expected nil for expired nonce on Peek")
	}
}

func TestSavePrunesExpired(t *testing.T) {
	s := New()

	// Insert an already-expired entry
	s.mu.Lock()
	s.entries["old"] = NonceEntry{
		AccessToken: "old-token",
		CreatedAt:   time.Now().Add(-31 * time.Second),
	}
	s.mu.Unlock()

	// Save a new entry — should prune the old one
	s.Save("new", NonceEntry{AccessToken: "new-token"})

	s.mu.Lock()
	_, oldExists := s.entries["old"]
	s.mu.Unlock()

	if oldExists {
		t.Error("expected expired entry to be pruned on Save")
	}
}
