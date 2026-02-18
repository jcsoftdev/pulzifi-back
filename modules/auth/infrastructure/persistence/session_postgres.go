package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
)

type SessionPostgresRepository struct {
	db *sql.DB
}

func NewSessionPostgresRepository(db *sql.DB) repositories.SessionRepository {
	return &SessionPostgresRepository{db: db}
}

func (r *SessionPostgresRepository) Create(ctx context.Context, session *entities.Session) error {
	query := `
		INSERT INTO public.sessions (id, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, query,
		session.ID,
		session.UserID,
		session.ExpiresAt,
		session.CreatedAt,
	)

	return err
}

func (r *SessionPostgresRepository) FindByID(ctx context.Context, id string) (*entities.Session, error) {
	query := `
		SELECT id, user_id, expires_at, created_at
		FROM public.sessions
		WHERE id = $1
	`

	var session entities.Session
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.ExpiresAt,
		&session.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &session, nil
}

func (r *SessionPostgresRepository) DeleteByID(ctx context.Context, id string) error {
	query := `
		DELETE FROM public.sessions
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *SessionPostgresRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM public.sessions
		WHERE expires_at < $1
	`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}
