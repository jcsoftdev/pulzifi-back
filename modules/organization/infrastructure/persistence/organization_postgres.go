package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/entities"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// OrganizationPostgresRepository implements OrganizationRepository using PostgreSQL
type OrganizationPostgresRepository struct {
	db *sql.DB
}

// NewOrganizationPostgresRepository creates a new PostgreSQL repository
func NewOrganizationPostgresRepository(db *sql.DB) *OrganizationPostgresRepository {
	return &OrganizationPostgresRepository{
		db: db,
	}
}

// Create stores a new organization
func (r *OrganizationPostgresRepository) Create(ctx context.Context, org *entities.Organization) error {
	query := `
		INSERT INTO public.organizations (id, name, subdomain, schema_name, owner_user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.ExecContext(ctx, query,
		org.ID,
		org.Name,
		org.Subdomain,
		org.SchemaName,
		org.OwnerUserID,
		org.CreatedAt,
		org.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to create organization", zap.Error(err))
		return err
	}

	return nil
}

// GetByID retrieves an organization by its ID
func (r *OrganizationPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Organization, error) {
	query := `
		SELECT id, name, subdomain, schema_name, owner_user_id, created_at, updated_at, deleted_at
		FROM public.organizations
		WHERE id = $1
	`

	var org entities.Organization
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.Subdomain,
		&org.SchemaName,
		&org.OwnerUserID,
		&org.CreatedAt,
		&org.UpdatedAt,
		&org.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Error("Failed to get organization by ID", zap.Error(err))
		return nil, err
	}

	return &org, nil
}

// GetBySubdomain retrieves an organization by its subdomain
func (r *OrganizationPostgresRepository) GetBySubdomain(ctx context.Context, subdomain string) (*entities.Organization, error) {
	query := `
		SELECT id, name, subdomain, schema_name, owner_user_id, created_at, updated_at, deleted_at
		FROM public.organizations
		WHERE subdomain = $1 AND deleted_at IS NULL
	`

	var org entities.Organization
	err := r.db.QueryRowContext(ctx, query, subdomain).Scan(
		&org.ID,
		&org.Name,
		&org.Subdomain,
		&org.SchemaName,
		&org.OwnerUserID,
		&org.CreatedAt,
		&org.UpdatedAt,
		&org.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Error("Failed to get organization by subdomain", zap.Error(err))
		return nil, err
	}

	return &org, nil
}

// List retrieves all organizations for a user (paginated)
func (r *OrganizationPostgresRepository) List(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.Organization, error) {
	query := `
		SELECT o.id, o.name, o.subdomain, o.schema_name, o.owner_user_id, o.created_at, o.updated_at, o.deleted_at
		FROM public.organizations o
		WHERE o.owner_user_id = $1 AND o.deleted_at IS NULL
		ORDER BY o.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		logger.Error("Failed to list organizations", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var organizations []*entities.Organization
	for rows.Next() {
		var org entities.Organization
		if err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.Subdomain,
			&org.SchemaName,
			&org.OwnerUserID,
			&org.CreatedAt,
			&org.UpdatedAt,
			&org.DeletedAt,
		); err != nil {
			logger.Error("Failed to scan organization row", zap.Error(err))
			return nil, err
		}
		organizations = append(organizations, &org)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error iterating organizations", zap.Error(err))
		return nil, err
	}

	return organizations, nil
}

// Update modifies an existing organization
func (r *OrganizationPostgresRepository) Update(ctx context.Context, org *entities.Organization) error {
	query := `
		UPDATE public.organizations
		SET name = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, org.Name, org.UpdatedAt, org.ID)
	if err != nil {
		logger.Error("Failed to update organization", zap.Error(err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", zap.Error(err))
		return err
	}

	if rowsAffected == 0 {
		return &domainerrors.OrganizationNotFoundError{OrganizationID: org.ID.String()}
	}

	return nil
}

// Delete soft-deletes an organization
func (r *OrganizationPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE public.organizations
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		logger.Error("Failed to delete organization", zap.Error(err))
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Failed to get rows affected", zap.Error(err))
		return err
	}

	if rowsAffected == 0 {
		return &domainerrors.OrganizationNotFoundError{OrganizationID: id.String()}
	}

	return nil
}

// CountBySubdomain checks if a subdomain already exists
func (r *OrganizationPostgresRepository) CountBySubdomain(ctx context.Context, subdomain string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM public.organizations
		WHERE subdomain = $1 AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, subdomain).Scan(&count)
	if err != nil {
		logger.Error("Failed to count organizations by subdomain", zap.Error(err))
		return 0, err
	}

	return count, nil
}
