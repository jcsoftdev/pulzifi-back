package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/lib/pq"
)

type NotificationPreferencePostgresRepository struct {
	db     *sql.DB
	tenant string
}

func NewNotificationPreferencePostgresRepository(db *sql.DB, tenant string) *NotificationPreferencePostgresRepository {
	return &NotificationPreferencePostgresRepository{db: db, tenant: tenant}
}

func (r *NotificationPreferencePostgresRepository) Create(ctx context.Context, pref *entities.NotificationPreference) error {
	q := `INSERT INTO notification_preferences (id, user_id, workspace_id, page_id, email_enabled, change_types, created_at, updated_at)
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q, pref.ID, pref.UserID, pref.WorkspaceID, pref.PageID, pref.EmailEnabled, pq.Array(pref.ChangeTypes), pref.CreatedAt, pref.UpdatedAt)
		return err
	})
}

func (r *NotificationPreferencePostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.NotificationPreference, error) {
	var p entities.NotificationPreference
	var changeTypes pq.StringArray
	q := `SELECT id, user_id, workspace_id, page_id, email_enabled, change_types, created_at, updated_at
		  FROM notification_preferences WHERE id = $1`

	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, q, id).Scan(&p.ID, &p.UserID, &p.WorkspaceID, &p.PageID, &p.EmailEnabled, &changeTypes, &p.CreatedAt, &p.UpdatedAt)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.ChangeTypes = changeTypes
	return &p, nil
}

func (r *NotificationPreferencePostgresRepository) GetByUserAndWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) (*entities.NotificationPreference, error) {
	var p entities.NotificationPreference
	var changeTypes pq.StringArray
	q := `SELECT id, user_id, workspace_id, page_id, email_enabled, change_types, created_at, updated_at
		  FROM notification_preferences WHERE user_id = $1 AND workspace_id = $2`

	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, q, userID, workspaceID).Scan(&p.ID, &p.UserID, &p.WorkspaceID, &p.PageID, &p.EmailEnabled, &changeTypes, &p.CreatedAt, &p.UpdatedAt)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.ChangeTypes = changeTypes
	return &p, nil
}

func (r *NotificationPreferencePostgresRepository) GetByUserAndPage(ctx context.Context, userID, pageID uuid.UUID) (*entities.NotificationPreference, error) {
	var p entities.NotificationPreference
	var changeTypes pq.StringArray
	q := `SELECT id, user_id, workspace_id, page_id, email_enabled, change_types, created_at, updated_at
		  FROM notification_preferences WHERE user_id = $1 AND page_id = $2`

	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, q, userID, pageID).Scan(&p.ID, &p.UserID, &p.WorkspaceID, &p.PageID, &p.EmailEnabled, &changeTypes, &p.CreatedAt, &p.UpdatedAt)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.ChangeTypes = changeTypes
	return &p, nil
}

func (r *NotificationPreferencePostgresRepository) Update(ctx context.Context, pref *entities.NotificationPreference) error {
	pref.UpdatedAt = time.Now()
	q := `UPDATE notification_preferences SET email_enabled = $1, change_types = $2, updated_at = $3 WHERE id = $4`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q, pref.EmailEnabled, pq.Array(pref.ChangeTypes), pref.UpdatedAt, pref.ID)
		return err
	})
}

func (r *NotificationPreferencePostgresRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM notification_preferences WHERE id = $1`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q, id)
		return err
	})
}

func (r *NotificationPreferencePostgresRepository) GetEmailEnabledByPage(ctx context.Context, pageID uuid.UUID) ([]*entities.NotificationPreference, error) {
	q := `SELECT id, user_id, workspace_id, page_id, email_enabled, change_types, created_at, updated_at
		  FROM notification_preferences WHERE page_id = $1 AND email_enabled = true`

	var prefs []*entities.NotificationPreference
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, q, pageID)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var p entities.NotificationPreference
			var changeTypes pq.StringArray
			if err := rows.Scan(&p.ID, &p.UserID, &p.WorkspaceID, &p.PageID, &p.EmailEnabled, &changeTypes, &p.CreatedAt, &p.UpdatedAt); err != nil {
				return err
			}
			p.ChangeTypes = changeTypes
			prefs = append(prefs, &p)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return prefs, nil
}
