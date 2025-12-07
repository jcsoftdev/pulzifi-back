package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/entities"
)

type AlertPostgresRepository struct {
	db     *sql.DB
	tenant string
}

func NewAlertPostgresRepository(db *sql.DB, tenant string) *AlertPostgresRepository {
	return &AlertPostgresRepository{db: db, tenant: tenant}
}

func (r *AlertPostgresRepository) Create(ctx context.Context, alert *entities.Alert) error {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return err
	}
	q := `INSERT INTO alerts (id, workspace_id, page_id, check_id, type, title, description, metadata, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.ExecContext(ctx, q, alert.ID, alert.WorkspaceID, alert.PageID, alert.CheckID, alert.Type, alert.Title, alert.Description, alert.Metadata, alert.CreatedAt)
	return err
}

func (r *AlertPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Alert, error) {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return nil, err
	}
	var a entities.Alert
	var readAt sql.NullTime
	q := `SELECT id, workspace_id, page_id, check_id, type, title, description, metadata, read_at, created_at FROM alerts WHERE id = $1`
	err := r.db.QueryRowContext(ctx, q, id).Scan(&a.ID, &a.WorkspaceID, &a.PageID, &a.CheckID, &a.Type, &a.Title, &a.Description, &a.Metadata, &readAt, &a.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if readAt.Valid {
		a.ReadAt = &readAt.Time
	}
	return &a, nil
}

func (r *AlertPostgresRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entities.Alert, error) {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return nil, err
	}
	q := `SELECT id, workspace_id, page_id, check_id, type, title, description, metadata, read_at, created_at FROM alerts WHERE workspace_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, q, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var alerts []*entities.Alert
	for rows.Next() {
		var a entities.Alert
		var readAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.WorkspaceID, &a.PageID, &a.CheckID, &a.Type, &a.Title, &a.Description, &a.Metadata, &readAt, &a.CreatedAt); err != nil {
			return nil, err
		}
		if readAt.Valid {
			a.ReadAt = &readAt.Time
		}
		alerts = append(alerts, &a)
	}
	return alerts, nil
}

func (r *AlertPostgresRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return err
	}
	q := `UPDATE alerts SET read_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, q, time.Now(), id)
	return err
}
