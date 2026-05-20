// Package teamwiring provides cross-module adapters for the team module.
// It bridges email and database token generation without creating direct
// module-to-module imports in the team module.
package teamwiring

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/templates"
	teamservices "github.com/jcsoftdev/pulzifi-back/modules/team/domain/services"
)

// --- EmailerAdapter ---

// EmailerAdapter adapts email.EmailProvider to team's Emailer interface.
type EmailerAdapter struct {
	provider emailservices.EmailProvider
}

// NewEmailerAdapter wraps an EmailProvider as team's Emailer.
func NewEmailerAdapter(p emailservices.EmailProvider) teamservices.Emailer {
	return &EmailerAdapter{provider: p}
}

func (a *EmailerAdapter) Send(ctx context.Context, to, subject, html string) error {
	return a.provider.Send(ctx, to, subject, html)
}

// --- InviteEmailBuilderAdapter ---

// InviteEmailBuilderAdapter uses the email module's template functions.
type InviteEmailBuilderAdapter struct{}

// NewInviteEmailBuilderAdapter creates an adapter for building team invite emails.
func NewInviteEmailBuilderAdapter() teamservices.InviteEmailBuilder {
	return &InviteEmailBuilderAdapter{}
}

func (a *InviteEmailBuilderAdapter) TeamInvite(inviterEmail, subdomain, loginURL string) (string, string) {
	return templates.TeamInvite(inviterEmail, subdomain, loginURL)
}

// --- TokenGeneratorAdapter ---

// TokenGeneratorAdapter generates password-reset tokens and stores them in the DB.
type TokenGeneratorAdapter struct {
	db *sql.DB
}

// NewTokenGeneratorAdapter creates a token generator backed by the given DB.
func NewTokenGeneratorAdapter(db *sql.DB) teamservices.TokenGenerator {
	return &TokenGeneratorAdapter{db: db}
}

func (a *TokenGeneratorAdapter) Generate(ctx context.Context, userID uuid.UUID) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(72 * time.Hour) // 3 days for invite tokens

	_, err := a.db.ExecContext(ctx,
		`INSERT INTO public.password_resets (id, user_id, token, expires_at, created_at) VALUES ($1, $2, $3, $4, NOW())`,
		uuid.New(), userID, token, expiresAt,
	)
	if err != nil {
		return "", err
	}
	return token, nil
}
