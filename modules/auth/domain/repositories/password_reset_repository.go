package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PasswordReset holds the data needed for a password reset operation.
type PasswordReset struct {
	UserID    uuid.UUID
	ExpiresAt time.Time
	Used      bool
}

// PasswordResetRepository manages password_resets in the public schema.
type PasswordResetRepository interface {
	// Store persists a new password reset token.
	Store(ctx context.Context, id uuid.UUID, userID uuid.UUID, token string, expiresAt time.Time) error

	// Find looks up a reset token and returns its metadata.
	// Returns nil if the token does not exist.
	Find(ctx context.Context, token string) (*PasswordReset, error)

	// Consume atomically marks the token as used, updates the user password,
	// and activates any pending organization memberships.
	Consume(ctx context.Context, token string, userID uuid.UUID, hashedPassword string) error
}
