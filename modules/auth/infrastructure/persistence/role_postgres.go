package persistence

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

type RolePostgresRepository struct {
	db *sql.DB
}

func NewRolePostgresRepository(db *sql.DB) *RolePostgresRepository {
	return &RolePostgresRepository{db: db}
}

func (r *RolePostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Role, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM public.roles WHERE id = $1`

	var role entities.Role
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logger.Error("Failed to get role by ID", zap.Error(err))
		return nil, err
	}

	return &role, nil
}

func (r *RolePostgresRepository) GetByName(ctx context.Context, name string) (*entities.Role, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM public.roles WHERE name = $1`

	var role entities.Role
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		logger.Error("Failed to get role by name", zap.Error(err))
		return nil, err
	}

	return &role, nil
}

func (r *RolePostgresRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*entities.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.created_at, r.updated_at
		FROM public.roles r
		INNER JOIN public.user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		logger.Error("Failed to get user roles", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var roles []*entities.Role
	for rows.Next() {
		var role entities.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt); err != nil {
			logger.Error("Failed to scan role", zap.Error(err))
			return nil, err
		}
		roles = append(roles, &role)
	}

	return roles, nil
}

func (r *RolePostgresRepository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]*entities.Permission, error) {
	query := `
		SELECT p.id, p.name, p.resource, p.action, p.description, p.created_at, p.updated_at
		FROM public.permissions p
		INNER JOIN public.role_permissions rp ON rp.permission_id = p.id
		WHERE rp.role_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		logger.Error("Failed to get role permissions", zap.Error(err))
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

func (r *RolePostgresRepository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	query := `INSERT INTO public.user_roles (user_id, role_id, created_at) VALUES ($1, $2, NOW())`

	_, err := r.db.ExecContext(ctx, query, userID, roleID)
	if err != nil {
		logger.Error("Failed to assign role to user", zap.Error(err))
		return err
	}

	return nil
}
