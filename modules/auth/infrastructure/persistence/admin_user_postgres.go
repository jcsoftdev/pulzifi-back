package persistence

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	listusers "github.com/jcsoftdev/pulzifi-back/modules/auth/application/list_users"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// AdminUserPostgresRepository implements list_users.AdminUserReader.
// All queries target the public schema — no SET search_path required.
type AdminUserPostgresRepository struct {
	db *sql.DB
}

// NewAdminUserPostgresRepository creates a new repository backed by db.
func NewAdminUserPostgresRepository(db *sql.DB) *AdminUserPostgresRepository {
	return &AdminUserPostgresRepository{db: db}
}

// ListUsers returns a paginated list of users with their SUPER_ADMIN status
// computed in a single SQL query to avoid N+1 role lookups.
func (r *AdminUserPostgresRepository) ListUsers(ctx context.Context, search string, limit, offset int) ([]listusers.AdminUserRow, int, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM public.users u
		WHERE u.deleted_at IS NULL
		  AND ($1 = '' OR u.email ILIKE '%' || $1 || '%'
		       OR (u.first_name || ' ' || u.last_name) ILIKE '%' || $1 || '%')
	`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, search).Scan(&total); err != nil {
		logger.Error("AdminUserRepo: count query failed", zap.Error(err))
		return nil, 0, err
	}

	const listQuery = `
		SELECT u.id,
		       u.email,
		       u.first_name,
		       u.last_name,
		       EXISTS(
		         SELECT 1
		         FROM public.user_roles ur
		         JOIN public.roles r ON r.id = ur.role_id
		         WHERE ur.user_id = u.id AND r.name = 'SUPER_ADMIN'
		       ) AS is_super_admin
		FROM public.users u
		WHERE u.deleted_at IS NULL
		  AND ($1 = '' OR u.email ILIKE '%' || $1 || '%'
		       OR (u.first_name || ' ' || u.last_name) ILIKE '%' || $1 || '%')
		ORDER BY u.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, listQuery, search, limit, offset)
	if err != nil {
		logger.Error("AdminUserRepo: list query failed", zap.Error(err))
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var result []listusers.AdminUserRow
	for rows.Next() {
		var row listusers.AdminUserRow
		var idStr uuid.UUID
		if err := rows.Scan(&idStr, &row.Email, &row.FirstName, &row.LastName, &row.IsSuperAdmin); err != nil {
			logger.Error("AdminUserRepo: scan failed", zap.Error(err))
			return nil, 0, err
		}
		row.ID = idStr
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return result, total, nil
}
