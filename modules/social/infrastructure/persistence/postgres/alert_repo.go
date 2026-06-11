package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/lib/pq"
)

// Compile-time interface check.
var _ repositories.AlertRepository = (*AlertPostgresRepository)(nil)

// AlertPostgresRepository implements repositories.AlertRepository using PostgreSQL.
type AlertPostgresRepository struct {
	db     *sql.DB
	tenant string
}

// NewAlertPostgresRepository creates a tenant-scoped AlertRepository.
func NewAlertPostgresRepository(db *sql.DB, tenant string) *AlertPostgresRepository {
	return &AlertPostgresRepository{db: db, tenant: tenant}
}

// Save persists a new SocialAlert.
func (r *AlertPostgresRepository) Save(ctx context.Context, alert *entities.SocialAlert) error {
	changeTypeStrs := make([]string, len(alert.ChangeTypes))
	for i, ct := range alert.ChangeTypes {
		changeTypeStrs[i] = string(ct)
	}

	q := `
		INSERT INTO social_alerts
			(id, profile_id, workspace_id, change_id, alert_type, title, description, change_types, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q,
			alert.ID,
			alert.ProfileID,
			alert.WorkspaceID,
			alert.ChangeID,
			alert.AlertType,
			alert.Title,
			alert.Description,
			pq.Array(changeTypeStrs),
			alert.IsRead,
			alert.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("alert repo: save: %w", err)
		}
		return nil
	})
}

// ListByWorkspace returns alerts for a workspace, newest first.
func (r *AlertPostgresRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*entities.SocialAlert, error) {
	if limit <= 0 {
		limit = 50
	}
	q := `
		SELECT id, profile_id, workspace_id, change_id, alert_type, title, description, change_types, is_read, created_at
		FROM social_alerts
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	var alerts []*entities.SocialAlert
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, q, workspaceID, limit, offset)
		if err != nil {
			return fmt.Errorf("alert repo: list: query: %w", err)
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var a entities.SocialAlert
			var changeTypeStrs []string
			var changeID uuid.NullUUID
			if err := rows.Scan(
				&a.ID, &a.ProfileID, &a.WorkspaceID, &changeID,
				&a.AlertType, &a.Title, &a.Description,
				pq.Array(&changeTypeStrs), &a.IsRead, &a.CreatedAt,
			); err != nil {
				return fmt.Errorf("alert repo: list: scan: %w", err)
			}
			if changeID.Valid {
				cid := changeID.UUID
				a.ChangeID = &cid
			}
			a.ChangeTypes = make([]valueobjects.ChangeType, len(changeTypeStrs))
			for i, s := range changeTypeStrs {
				a.ChangeTypes[i] = valueobjects.ChangeType(s)
			}
			alerts = append(alerts, &a)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return alerts, nil
}

// MarkRead flips a single alert's is_read to true.
func (r *AlertPostgresRepository) MarkRead(ctx context.Context, alertID uuid.UUID) error {
	q := `UPDATE social_alerts SET is_read = TRUE WHERE id = $1`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q, alertID)
		if err != nil {
			return fmt.Errorf("alert repo: mark read: %w", err)
		}
		return nil
	})
}

// UnreadCount returns the number of unread alerts for a workspace.
func (r *AlertPostgresRepository) UnreadCount(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	q := `SELECT COUNT(*) FROM social_alerts WHERE workspace_id = $1 AND is_read = FALSE`
	var count int
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, q, workspaceID).Scan(&count)
	})
	if err != nil {
		return 0, fmt.Errorf("alert repo: unread count: %w", err)
	}
	return count, nil
}
