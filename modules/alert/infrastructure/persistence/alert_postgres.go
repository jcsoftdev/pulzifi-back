package persistence

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
)

// dbMetadata wraps entities.Metadata to provide sql.Scanner and driver.Valuer
// without polluting the domain layer with infrastructure imports.
type dbMetadata entities.Metadata

func (m dbMetadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *dbMetadata) Scan(value interface{}) error {
	if value == nil {
		*m = make(dbMetadata)
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, m)
}

type AlertPostgresRepository struct {
	db     *sql.DB
	tenant string
}

func NewAlertPostgresRepository(db *sql.DB, tenant string) *AlertPostgresRepository {
	return &AlertPostgresRepository{db: db, tenant: tenant}
}

func (r *AlertPostgresRepository) Create(ctx context.Context, alert *entities.Alert) error {
	q := `INSERT INTO alerts (id, workspace_id, page_id, check_id, type, title, description, change_summary, metadata, created_at)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		WHERE EXISTS (SELECT 1 FROM checks WHERE id = $4)`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q, alert.ID, alert.WorkspaceID, alert.PageID, alert.CheckID, alert.Type, alert.Title, alert.Description, alert.ChangeSummary, dbMetadata(alert.Metadata), alert.CreatedAt)
		return err
	})
}

func (r *AlertPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Alert, error) {
	var a entities.Alert
	var readAt sql.NullTime
	var metadata dbMetadata
	q := `SELECT id, workspace_id, page_id, check_id, type, title, description, COALESCE(change_summary, ''), metadata, read_at, created_at FROM alerts WHERE id = $1`

	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, q, id).Scan(&a.ID, &a.WorkspaceID, &a.PageID, &a.CheckID, &a.Type, &a.Title, &a.Description, &a.ChangeSummary, &metadata, &readAt, &a.CreatedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return sql.ErrNoRows
			}
			return err
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	a.Metadata = entities.Metadata(metadata)
	if readAt.Valid {
		a.ReadAt = &readAt.Time
	}
	return &a, nil
}

func (r *AlertPostgresRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entities.Alert, error) {
	q := `SELECT id, workspace_id, page_id, check_id, type, title, description, COALESCE(change_summary, ''), metadata, read_at, created_at FROM alerts WHERE workspace_id = $1 ORDER BY created_at DESC`

	var alerts []*entities.Alert
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, q, workspaceID)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var a entities.Alert
			var readAt sql.NullTime
			var metadata dbMetadata
			if err := rows.Scan(&a.ID, &a.WorkspaceID, &a.PageID, &a.CheckID, &a.Type, &a.Title, &a.Description, &a.ChangeSummary, &metadata, &readAt, &a.CreatedAt); err != nil {
				return err
			}
			a.Metadata = entities.Metadata(metadata)
			if readAt.Valid {
				a.ReadAt = &readAt.Time
			}
			alerts = append(alerts, &a)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return alerts, nil
}

func (r *AlertPostgresRepository) CountUnread(ctx context.Context) (int, error) {
	var count int
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts WHERE read_at IS NULL`).Scan(&count)
	})
	return count, err
}

func (r *AlertPostgresRepository) CountAll(ctx context.Context) (int, error) {
	var count int
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		return tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM alerts`).Scan(&count)
	})
	return count, err
}

func (r *AlertPostgresRepository) ListAll(ctx context.Context, limit int) ([]*entities.AlertWithPage, error) {
	q := `SELECT a.id, a.workspace_id, a.page_id, a.check_id, a.type, a.title, a.description, a.metadata, a.read_at, a.created_at,
	       COALESCE(p.name, '') AS page_name, COALESCE(p.url, '') AS page_url
	       FROM alerts a LEFT JOIN pages p ON p.id = a.page_id
	       ORDER BY a.created_at DESC LIMIT $1`

	var alerts []*entities.AlertWithPage
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, q, limit)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var a entities.AlertWithPage
			var readAt sql.NullTime
			var metadata dbMetadata
			if err := rows.Scan(&a.ID, &a.WorkspaceID, &a.PageID, &a.CheckID, &a.Type, &a.Title, &a.Description, &metadata, &readAt, &a.CreatedAt, &a.PageName, &a.PageURL); err != nil {
				return err
			}
			a.Metadata = entities.Metadata(metadata)
			if readAt.Valid {
				a.ReadAt = &readAt.Time
			}
			alerts = append(alerts, &a)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return alerts, nil
}

func (r *AlertPostgresRepository) MarkAllAsRead(ctx context.Context) error {
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `UPDATE alerts SET read_at = $1 WHERE read_at IS NULL`, time.Now())
		return err
	})
}

func (r *AlertPostgresRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE alerts SET read_at = $1 WHERE id = $2`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q, time.Now(), id)
		return err
	})
}

func (r *AlertPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `DELETE FROM alerts WHERE id = $1`, id)
		return err
	})
}
