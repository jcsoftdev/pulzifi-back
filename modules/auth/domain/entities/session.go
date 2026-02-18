package entities

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        string
	UserID    uuid.UUID
	ExpiresAt time.Time
	CreatedAt time.Time
}

func NewSession(userID uuid.UUID, ttl time.Duration) *Session {
	now := time.Now()
	return &Session{
		ID:        uuid.NewString(),
		UserID:    userID,
		ExpiresAt: now.Add(ttl),
		CreatedAt: now,
	}
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
