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

// ListUsers returns a paginated list of users with their SUPER_ADMIN status,
// status, email_verified, and org_count computed in a single SQL query.
func (r *AdminUserPostgresRepository) ListUsers(ctx context.Context, filter listusers.ListFilter, limit, offset int) ([]listusers.AdminUserRow, int, error) {
	// $1 search, $2 orgID ('' = no filter), $3 status ('' = no filter)
	const where = `
		WHERE u.deleted_at IS NULL
		  AND ($1 = '' OR u.email ILIKE '%' || $1 || '%'
		       OR (u.first_name || ' ' || u.last_name) ILIKE '%' || $1 || '%')
		  AND ($3 = '' OR u.status = $3)
		  AND ($2 = '' OR EXISTS (
		        SELECT 1 FROM public.organization_members om
		        WHERE om.user_id = u.id AND om.deleted_at IS NULL
		          AND om.organization_id::text = $2))
	`

	countQuery := `SELECT COUNT(*) FROM public.users u ` + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, filter.Search, filter.OrgID, filter.Status).Scan(&total); err != nil {
		logger.Error("AdminUserRepo: count query failed", zap.Error(err))
		return nil, 0, err
	}

	listQuery := `
		SELECT u.id, u.email, u.first_name, u.last_name, u.status, u.email_verified,
		       EXISTS(
		         SELECT 1 FROM public.user_roles ur
		         JOIN public.roles ro ON ro.id = ur.role_id
		         WHERE ur.user_id = u.id AND ro.name = 'SUPER_ADMIN'
		       ) AS is_super_admin,
		       (SELECT COUNT(*) FROM public.organization_members om2
		          WHERE om2.user_id = u.id AND om2.deleted_at IS NULL) AS org_count
		FROM public.users u ` + where + `
		ORDER BY u.created_at DESC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.db.QueryContext(ctx, listQuery, filter.Search, filter.OrgID, filter.Status, limit, offset)
	if err != nil {
		logger.Error("AdminUserRepo: list query failed", zap.Error(err))
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var result []listusers.AdminUserRow
	for rows.Next() {
		var row listusers.AdminUserRow
		if err := rows.Scan(&row.ID, &row.Email, &row.FirstName, &row.LastName,
			&row.Status, &row.EmailVerified, &row.IsSuperAdmin, &row.OrgCount); err != nil {
			logger.Error("AdminUserRepo: scan failed", zap.Error(err))
			return nil, 0, err
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return result, total, nil
}

// Ensure the uuid import is used (ID is scanned directly by the driver).
var _ = uuid.UUID{}
