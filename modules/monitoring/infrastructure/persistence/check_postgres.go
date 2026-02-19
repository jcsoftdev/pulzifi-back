package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
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
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}

	q := `INSERT INTO checks (id, page_id, status, screenshot_url, html_snapshot_url, content_hash, change_detected, change_type, error_message, duration_ms, checked_at) 
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.ExecContext(ctx, q,
		check.ID,
		check.PageID,
		check.Status,
		check.ScreenshotURL,
		check.HTMLSnapshotURL,
		check.ContentHash,
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
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	var check entities.Check
	q := `SELECT id, page_id, status, COALESCE(screenshot_url, ''), COALESCE(html_snapshot_url, ''), COALESCE(content_hash, ''), COALESCE(change_detected, false), COALESCE(change_type, ''), COALESCE(error_message, ''), COALESCE(duration_ms, 0), checked_at 
	      FROM checks WHERE id = $1`

	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&check.ID,
		&check.PageID,
		&check.Status,
		&check.ScreenshotURL,
		&check.HTMLSnapshotURL,
		&check.ContentHash,
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

// Update updates an existing check
func (r *CheckPostgresRepository) Update(ctx context.Context, check *entities.Check) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}

	q := `UPDATE checks SET 
		status = $1, 
		screenshot_url = $2, 
		html_snapshot_url = $3, 
		content_hash = $4, 
		change_detected = $5, 
		change_type = $6, 
		error_message = $7, 
		duration_ms = $8 
		WHERE id = $9`

	_, err := r.db.ExecContext(ctx, q,
		check.Status,
		check.ScreenshotURL,
		check.HTMLSnapshotURL,
		check.ContentHash,
		check.ChangeDetected,
		check.ChangeType,
		check.ErrorMessage,
		check.DurationMs,
		check.ID,
	)
	return err
}

// ListByPage retrieves all checks for a page
func (r *CheckPostgresRepository) ListByPage(ctx context.Context, pageID uuid.UUID) ([]*entities.Check, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	q := `SELECT id, page_id, status, COALESCE(screenshot_url, ''), COALESCE(html_snapshot_url, ''), COALESCE(content_hash, ''), COALESCE(change_detected, false), COALESCE(change_type, ''), COALESCE(error_message, ''), COALESCE(duration_ms, 0), checked_at 
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
			&check.ContentHash,
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

// GetPreviousSuccessfulByPage retrieves the most recent successful check for a page
// excluding the provided check ID (used to compare against the current check).
func (r *CheckPostgresRepository) GetPreviousSuccessfulByPage(ctx context.Context, pageID, excludeCheckID uuid.UUID) (*entities.Check, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	var check entities.Check
	q := `SELECT id, page_id, status, COALESCE(screenshot_url, ''), COALESCE(html_snapshot_url, ''), COALESCE(content_hash, ''), COALESCE(change_detected, false), COALESCE(change_type, ''), COALESCE(error_message, ''), COALESCE(duration_ms, 0), checked_at
	      FROM checks WHERE page_id = $1 AND id != $2 AND status = 'success' ORDER BY checked_at DESC LIMIT 1`

	err := r.db.QueryRowContext(ctx, q, pageID, excludeCheckID).Scan(
		&check.ID,
		&check.PageID,
		&check.Status,
		&check.ScreenshotURL,
		&check.HTMLSnapshotURL,
		&check.ContentHash,
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

// GetLatestByPage retrieves the latest check for a page
func (r *CheckPostgresRepository) GetLatestByPage(ctx context.Context, pageID uuid.UUID) (*entities.Check, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	var check entities.Check
	q := `SELECT id, page_id, status, COALESCE(screenshot_url, ''), COALESCE(html_snapshot_url, ''), COALESCE(content_hash, ''), COALESCE(change_detected, false), COALESCE(change_type, ''), COALESCE(error_message, ''), COALESCE(duration_ms, 0), checked_at 
	      FROM checks WHERE page_id = $1 ORDER BY checked_at DESC LIMIT 1`

	err := r.db.QueryRowContext(ctx, q, pageID).Scan(
		&check.ID,
		&check.PageID,
		&check.Status,
		&check.ScreenshotURL,
		&check.HTMLSnapshotURL,
		&check.ContentHash,
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
