package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// UserPostgresRepository implements UserRepository using PostgreSQL
type UserPostgresRepository struct {
	db *sql.DB
}

// NewUserPostgresRepository creates a new PostgreSQL repository
func NewUserPostgresRepository(db *sql.DB) *UserPostgresRepository {
	return &UserPostgresRepository{
		db: db,
	}
}

// Create stores a new user
func (r *UserPostgresRepository) Create(ctx context.Context, user *entities.User) error {
	query := `
		INSERT INTO public.users (id, email, password_hash, first_name, last_name, avatar_url, email_verified, email_notifications_enabled, notification_frequency, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.AvatarURL,
		user.EmailVerified,
		user.EmailNotificationsEnabled,
		user.NotificationFrequency,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return err
	}

	return nil
}

// GetByID retrieves a user by their ID
func (r *UserPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, avatar_url, email_verified, email_notifications_enabled, notification_frequency, created_at, updated_at, deleted_at
		FROM public.users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user entities.User
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.EmailVerified,
		&user.EmailNotificationsEnabled,
		&user.NotificationFrequency,
		&user.CreatedAt,
		&user.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Error("Failed to get user by ID", zap.Error(err))
		return nil, err
	}

	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return &user, nil
}

// GetByEmail retrieves a user by their email
func (r *UserPostgresRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, avatar_url, email_verified, email_notifications_enabled, notification_frequency, created_at, updated_at, deleted_at
		FROM public.users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user entities.User
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.EmailVerified,
		&user.EmailNotificationsEnabled,
		&user.NotificationFrequency,
		&user.CreatedAt,
		&user.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Error("Failed to get user by email", zap.Error(err))
		return nil, err
	}

	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return &user, nil
}

// Update modifies an existing user
func (r *UserPostgresRepository) Update(ctx context.Context, user *entities.User) error {
	query := `
		UPDATE public.users
		SET email = $1, password_hash = $2, first_name = $3, last_name = $4, avatar_url = $5,
		    email_verified = $6, email_notifications_enabled = $7, notification_frequency = $8, updated_at = $9
		WHERE id = $10
	`

	user.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.AvatarURL,
		user.EmailVerified,
		user.EmailNotificationsEnabled,
		user.NotificationFrequency,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		return err
	}

	return nil
}

// Delete soft-deletes a user
func (r *UserPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE public.users
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", zap.Error(err))
		return err
	}

	if rowsAffected == 0 {
		logger.Warn("User not found for deletion", zap.String("id", id.String()))
		return nil
	}

	return nil
}

// ExistsByEmail checks if a user with the given email exists
func (r *UserPostgresRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM public.users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, email).Scan(&count)
	if err != nil {
		logger.Error("Failed to check if user exists by email", zap.Error(err))
		return false, err
	}

	return count > 0, nil
}
