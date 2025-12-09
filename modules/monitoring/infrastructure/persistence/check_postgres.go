package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
)

// CheckPostgresRepository implements CheckRepository using PostgreSQL
type CheckPostgresRepository struct {
	db     *sql.DB
	tenant string
}

// NewCheckPostgresRepository creates a new PostgreSQL repository
func NewCheckPostgresRepository(db *sql.DB, tenant string) *CheckPostgresRepository {
	return &CheckPostgresRepository{
		db:     db,
		tenant: tenant,
	}
}

// Create stores a new check
func (r *CheckPostgresRepository) Create(ctx context.Context, check *entities.Check) error {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return err
	}

	q := `INSERT INTO checks (id, page_id, status, screenshot_url, html_snapshot_url, change_detected, change_type, error_message, duration_ms, checked_at) 
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(ctx, q,
		check.ID,
		check.PageID,
		check.Status,
		check.ScreenshotURL,
		check.HTMLSnapshotURL,
		check.ChangeDetected,
		check.ChangeType,
		check.ErrorMessage,
		check.DurationMs,
		check.CheckedAt,
	)
	return err
}

// GetByID retrieves a check by ID
func (r *CheckPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Check, error) {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return nil, err
	}

	var check entities.Check
	q := `SELECT id, page_id, status, screenshot_url, html_snapshot_url, change_detected, change_type, error_message, duration_ms, checked_at 
	      FROM checks WHERE id = $1`

	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&check.ID,
		&check.PageID,
		&check.Status,
		&check.ScreenshotURL,
		&check.HTMLSnapshotURL,
		&check.ChangeDetected,
		&check.ChangeType,
		&check.ErrorMessage,
		&check.DurationMs,
		&check.CheckedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &check, nil
}

// ListByPage retrieves all checks for a page
func (r *CheckPostgresRepository) ListByPage(ctx context.Context, pageID uuid.UUID) ([]*entities.Check, error) {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return nil, err
	}

	q := `SELECT id, page_id, status, screenshot_url, html_snapshot_url, change_detected, change_type, error_message, duration_ms, checked_at 
	      FROM checks WHERE page_id = $1 ORDER BY checked_at DESC`

	rows, err := r.db.QueryContext(ctx, q, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []*entities.Check
	for rows.Next() {
		var check entities.Check
		if err := rows.Scan(
			&check.ID,
			&check.PageID,
			&check.Status,
			&check.ScreenshotURL,
			&check.HTMLSnapshotURL,
			&check.ChangeDetected,
			&check.ChangeType,
			&check.ErrorMessage,
			&check.DurationMs,
			&check.CheckedAt,
		); err != nil {
			return nil, err
		}
		checks = append(checks, &check)
	}

	return checks, nil
}
