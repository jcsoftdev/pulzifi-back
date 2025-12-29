package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
)

type RefreshTokenPostgresRepository struct {
	db *sql.DB
}

func NewRefreshTokenPostgresRepository(db *sql.DB) repositories.RefreshTokenRepository {
	return &RefreshTokenPostgresRepository{db: db}
}

func (r *RefreshTokenPostgresRepository) Create(ctx context.Context, refreshToken *entities.RefreshToken) error {
	query := `
		INSERT INTO public.refresh_tokens (id, user_id, token, expires_at, is_revoked, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		refreshToken.ID,
		refreshToken.UserID,
		refreshToken.Token,
		refreshToken.ExpiresAt,
		refreshToken.IsRevoked,
		refreshToken.CreatedAt,
		refreshToken.UpdatedAt,
	)

	return err
}

func (r *RefreshTokenPostgresRepository) FindByToken(ctx context.Context, token string) (*entities.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, is_revoked, created_at, updated_at
		FROM public.refresh_tokens
		WHERE token = $1
	`

	var rt entities.RefreshToken
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.IsRevoked,
		&rt.CreatedAt,
		&rt.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("refresh token not found")
		}
		return nil, err
	}

	return &rt, nil
}

func (r *RefreshTokenPostgresRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, is_revoked, created_at, updated_at
		FROM public.refresh_tokens
		WHERE user_id = $1 AND is_revoked = false
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*entities.RefreshToken
	for rows.Next() {
		var rt entities.RefreshToken
		if err := rows.Scan(
			&rt.ID,
			&rt.UserID,
			&rt.Token,
			&rt.ExpiresAt,
			&rt.IsRevoked,
			&rt.CreatedAt,
			&rt.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tokens = append(tokens, &rt)
	}

	return tokens, rows.Err()
}

func (r *RefreshTokenPostgresRepository) Revoke(ctx context.Context, token string) error {
	query := `
		UPDATE public.refresh_tokens
		SET is_revoked = true, updated_at = $1
		WHERE token = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), token)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("refresh token not found")
	}

	return nil
}

func (r *RefreshTokenPostgresRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE public.refresh_tokens
		SET is_revoked = true, updated_at = $1
		WHERE user_id = $2 AND is_revoked = false
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

func (r *RefreshTokenPostgresRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM public.refresh_tokens
		WHERE expires_at < $1
	`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}
