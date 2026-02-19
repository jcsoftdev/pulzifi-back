package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// RegistrationRequestPostgresRepository implements RegistrationRequestRepository using PostgreSQL
type RegistrationRequestPostgresRepository struct {
	db *sql.DB
}

// NewRegistrationRequestPostgresRepository creates a new PostgreSQL repository
func NewRegistrationRequestPostgresRepository(db *sql.DB) *RegistrationRequestPostgresRepository {
	return &RegistrationRequestPostgresRepository{db: db}
}

// Create stores a new registration request
func (r *RegistrationRequestPostgresRepository) Create(ctx context.Context, req *entities.RegistrationRequest) error {
	query := `
		INSERT INTO public.registration_requests (id, user_id, organization_name, organization_subdomain, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		req.ID,
		req.UserID,
		req.OrganizationName,
		req.OrganizationSubdomain,
		req.Status,
		req.CreatedAt,
		req.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to create registration request", zap.Error(err))
		return err
	}

	return nil
}

// GetByID retrieves a registration request by its ID
func (r *RegistrationRequestPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.RegistrationRequest, error) {
	query := `
		SELECT id, user_id, organization_name, organization_subdomain, status, reviewed_by, reviewed_at, created_at, updated_at
		FROM public.registration_requests
		WHERE id = $1
	`

	return r.scanOne(ctx, query, id)
}

// GetByUserID retrieves a registration request by user ID
func (r *RegistrationRequestPostgresRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entities.RegistrationRequest, error) {
	query := `
		SELECT id, user_id, organization_name, organization_subdomain, status, reviewed_by, reviewed_at, created_at, updated_at
		FROM public.registration_requests
		WHERE user_id = $1
	`

	return r.scanOne(ctx, query, userID)
}

// ListPending retrieves all pending registration requests
func (r *RegistrationRequestPostgresRepository) ListPending(ctx context.Context, limit, offset int) ([]*entities.RegistrationRequest, error) {
	query := `
		SELECT id, user_id, organization_name, organization_subdomain, status, reviewed_by, reviewed_at, created_at, updated_at
		FROM public.registration_requests
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		logger.Error("Failed to list pending registration requests", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var requests []*entities.RegistrationRequest
	for rows.Next() {
		req, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// UpdateStatus updates the status of a registration request
func (r *RegistrationRequestPostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, reviewedBy uuid.UUID) error {
	query := `
		UPDATE public.registration_requests
		SET status = $1, reviewed_by = $2, reviewed_at = $3, updated_at = $3
		WHERE id = $4
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, status, reviewedBy, now, id)
	if err != nil {
		logger.Error("Failed to update registration request status", zap.Error(err))
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ExistsPendingBySubdomain returns true if there is already a pending request for the given subdomain
func (r *RegistrationRequestPostgresRepository) ExistsPendingBySubdomain(ctx context.Context, subdomain string) (bool, error) {
	query := `
		SELECT COUNT(*) FROM public.registration_requests
		WHERE organization_subdomain = $1 AND status = 'pending'
	`
	var count int
	if err := r.db.QueryRowContext(ctx, query, subdomain).Scan(&count); err != nil {
		logger.Error("Failed to check pending subdomain", zap.Error(err))
		return false, err
	}
	return count > 0, nil
}

func (r *RegistrationRequestPostgresRepository) scanOne(ctx context.Context, query string, arg interface{}) (*entities.RegistrationRequest, error) {
	var req entities.RegistrationRequest
	var reviewedBy sql.NullString
	var reviewedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&req.ID,
		&req.UserID,
		&req.OrganizationName,
		&req.OrganizationSubdomain,
		&req.Status,
		&reviewedBy,
		&reviewedAt,
		&req.CreatedAt,
		&req.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Error("Failed to scan registration request", zap.Error(err))
		return nil, err
	}

	if reviewedBy.Valid {
		id, _ := uuid.Parse(reviewedBy.String)
		req.ReviewedBy = &id
	}
	if reviewedAt.Valid {
		req.ReviewedAt = &reviewedAt.Time
	}

	return &req, nil
}

func (r *RegistrationRequestPostgresRepository) scanRow(rows *sql.Rows) (*entities.RegistrationRequest, error) {
	var req entities.RegistrationRequest
	var reviewedBy sql.NullString
	var reviewedAt sql.NullTime

	if err := rows.Scan(
		&req.ID,
		&req.UserID,
		&req.OrganizationName,
		&req.OrganizationSubdomain,
		&req.Status,
		&reviewedBy,
		&reviewedAt,
		&req.CreatedAt,
		&req.UpdatedAt,
	); err != nil {
		logger.Error("Failed to scan registration request row", zap.Error(err))
		return nil, err
	}

	if reviewedBy.Valid {
		id, _ := uuid.Parse(reviewedBy.String)
		req.ReviewedBy = &id
	}
	if reviewedAt.Valid {
		req.ReviewedAt = &reviewedAt.Time
	}

	return &req, nil
}
