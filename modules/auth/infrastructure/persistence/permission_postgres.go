package persistence

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type PermissionPostgresRepository struct {
	db *sql.DB
}

func NewPermissionPostgresRepository(db *sql.DB) *PermissionPostgresRepository {
	return &PermissionPostgresRepository{db: db}
}

func (r *PermissionPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Permission, error) {
	query := `SELECT id, name, resource, action, description, created_at, updated_at FROM public.permissions WHERE id = $1`

	var perm entities.Permission
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt, &perm.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logger.Error("Failed to get permission by ID", zap.Error(err))
		return nil, err
	}

	return &perm, nil
}

func (r *PermissionPostgresRepository) GetByName(ctx context.Context, name string) (*entities.Permission, error) {
	query := `SELECT id, name, resource, action, description, created_at, updated_at FROM public.permissions WHERE name = $1`

	var perm entities.Permission
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt, &perm.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logger.Error("Failed to get permission by name", zap.Error(err))
		return nil, err
	}

	return &perm, nil
}

func (r *PermissionPostgresRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*entities.Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.resource, p.action, p.description, p.created_at, p.updated_at
		FROM public.permissions p
		INNER JOIN public.role_permissions rp ON rp.permission_id = p.id
		INNER JOIN public.user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Error("Failed to get user permissions", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var permissions []*entities.Permission
	for rows.Next() {
		var perm entities.Permission
		if err := rows.Scan(&perm.ID, &perm.Name, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt, &perm.UpdatedAt); err != nil {
			logger.Error("Failed to scan permission", zap.Error(err))
			return nil, err
		}
		permissions = append(permissions, &perm)
	}

	return permissions, nil
}

func (r *PermissionPostgresRepository) HasPermission(ctx context.Context, userID uuid.UUID, resource, action string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM public.permissions p
			INNER JOIN public.role_permissions rp ON rp.permission_id = p.id
			INNER JOIN public.user_roles ur ON ur.role_id = rp.role_id
			WHERE ur.user_id = $1 AND p.resource = $2 AND p.action = $3
		)
	`

	var hasPermission bool
	err := r.db.QueryRowContext(ctx, query, userID, resource, action).Scan(&hasPermission)
	if err != nil {
		logger.Error("Failed to check permission", zap.Error(err))
		return false, err
	}

	return hasPermission, nil
}
