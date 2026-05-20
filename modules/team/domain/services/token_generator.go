package services

import (
	"context"

	"github.com/google/uuid"
)

// TokenGenerator creates password-reset / invite tokens and persists them.
// The concrete implementation uses a SQL database; the interface keeps the
// application layer free of database/sql imports.
type TokenGenerator interface {
	Generate(ctx context.Context, userID uuid.UUID) (token string, err error)
}
