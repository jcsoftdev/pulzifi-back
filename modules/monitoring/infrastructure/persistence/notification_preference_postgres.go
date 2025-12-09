package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
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
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return err
	}

	q := `INSERT INTO notification_preferences (id, user_id, workspace_id, page_id, email_enabled, change_types, created_at, updated_at) 
		  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, q, pref.ID, pref.UserID, pref.WorkspaceID, pref.PageID, pref.EmailEnabled, pq.Array(pref.ChangeTypes), pref.CreatedAt, pref.UpdatedAt)
	return err
}

func (r *NotificationPreferencePostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.NotificationPreference, error) {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return nil, err
	}

	var p entities.NotificationPreference
	var changeTypes pq.StringArray
	q := `SELECT id, user_id, workspace_id, page_id, email_enabled, change_types, created_at, updated_at 
		  FROM notification_preferences WHERE id = $1`
	err := r.db.QueryRowContext(ctx, q, id).Scan(&p.ID, &p.UserID, &p.WorkspaceID, &p.PageID, &p.EmailEnabled, &changeTypes, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	p.ChangeTypes = changeTypes
	return &p, nil
}

func (r *NotificationPreferencePostgresRepository) GetByUserAndWorkspace(ctx context.Context, userID, workspaceID uuid.UUID) (*entities.NotificationPreference, error) {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return nil, err
	}

	var p entities.NotificationPreference
	var changeTypes pq.StringArray
	q := `SELECT id, user_id, workspace_id, page_id, email_enabled, change_types, created_at, updated_at 
		  FROM notification_preferences WHERE user_id = $1 AND workspace_id = $2`
	err := r.db.QueryRowContext(ctx, q, userID, workspaceID).Scan(&p.ID, &p.UserID, &p.WorkspaceID, &p.PageID, &p.EmailEnabled, &changeTypes, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	p.ChangeTypes = changeTypes
	return &p, nil
}

func (r *NotificationPreferencePostgresRepository) GetByUserAndPage(ctx context.Context, userID, pageID uuid.UUID) (*entities.NotificationPreference, error) {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return nil, err
	}

	var p entities.NotificationPreference
	var changeTypes pq.StringArray
	q := `SELECT id, user_id, workspace_id, page_id, email_enabled, change_types, created_at, updated_at 
		  FROM notification_preferences WHERE user_id = $1 AND page_id = $2`
	err := r.db.QueryRowContext(ctx, q, userID, pageID).Scan(&p.ID, &p.UserID, &p.WorkspaceID, &p.PageID, &p.EmailEnabled, &changeTypes, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	p.ChangeTypes = changeTypes
	return &p, nil
}

func (r *NotificationPreferencePostgresRepository) Update(ctx context.Context, pref *entities.NotificationPreference) error {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return err
	}

	pref.UpdatedAt = time.Now()
	q := `UPDATE notification_preferences SET email_enabled = $1, change_types = $2, updated_at = $3 WHERE id = $4`
	_, err := r.db.ExecContext(ctx, q, pref.EmailEnabled, pq.Array(pref.ChangeTypes), pref.UpdatedAt, pref.ID)
	return err
}

func (r *NotificationPreferencePostgresRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return err
	}

	q := `DELETE FROM notification_preferences WHERE id = $1`
	_, err := r.db.ExecContext(ctx, q, id)
	return err
}
