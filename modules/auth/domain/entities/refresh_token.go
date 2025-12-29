package entities

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	IsRevoked bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewRefreshToken(userID uuid.UUID, token string, expiresAt time.Time) *RefreshToken {
	now := time.Now()
	return &RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		IsRevoked: false,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

func (rt *RefreshToken) IsValid() bool {
	return !rt.IsRevoked && !rt.IsExpired()
}

func (rt *RefreshToken) Revoke() {
	rt.IsRevoked = true
	rt.UpdatedAt = time.Now()
}
