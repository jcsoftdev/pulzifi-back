package persistence

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
)

// PasswordResetPostgresRepository implements repositories.PasswordResetRepository.
type PasswordResetPostgresRepository struct {
	db *sql.DB
}

// NewPasswordResetPostgresRepository creates a new password reset repository.
func NewPasswordResetPostgresRepository(db *sql.DB) repositories.PasswordResetRepository {
	return &PasswordResetPostgresRepository{db: db}
}

func (r *PasswordResetPostgresRepository) Store(ctx context.Context, id uuid.UUID, userID uuid.UUID, token string, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO public.password_resets (id, user_id, token, expires_at, created_at) VALUES ($1, $2, $3, $4, NOW())`,
		id, userID, token, expiresAt,
	)
	return err
}

func (r *PasswordResetPostgresRepository) Find(ctx context.Context, token string) (*repositories.PasswordReset, error) {
	var pr repositories.PasswordReset
	err := r.db.QueryRowContext(ctx,
		`SELECT user_id, expires_at, used FROM public.password_resets WHERE token = $1`,
		token,
	).Scan(&pr.UserID, &pr.ExpiresAt, &pr.Used)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *PasswordResetPostgresRepository) Consume(ctx context.Context, token string, userID uuid.UUID, hashedPassword string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	_, err = tx.ExecContext(ctx,
		`UPDATE public.users SET password_hash = $1, updated_at = NOW() WHERE id = $2`,
		hashedPassword, userID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE public.password_resets SET used = true WHERE token = $1`,
		token,
	)
	if err != nil {
		return err
	}

	// Activate pending org memberships (non-fatal on failure — proceed anyway)
	tx.ExecContext(ctx, //nolint:errcheck
		`UPDATE public.organization_members SET invitation_status = 'active' WHERE user_id = $1 AND invitation_status = 'pending'`,
		userID,
	)

	return tx.Commit()
}
