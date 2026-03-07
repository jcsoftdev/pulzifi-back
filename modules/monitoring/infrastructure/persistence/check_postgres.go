package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
)

const checkSelectColumns = `id, page_id, section_id, parent_check_id, status, COALESCE(screenshot_url, ''), COALESCE(html_snapshot_url, ''), COALESCE(content_hash, ''), COALESCE(change_detected, false), COALESCE(change_type, ''), COALESCE(error_message, ''), COALESCE(duration_ms, 0), COALESCE(screenshot_hash, ''), COALESCE(vision_change_summary, ''), checked_at`

func scanCheck(row interface{ Scan(...interface{}) error }, check *entities.Check) error {
	return row.Scan(
		&check.ID,
		&check.PageID,
		&check.SectionID,
		&check.ParentCheckID,
		&check.Status,
		&check.ScreenshotURL,
		&check.HTMLSnapshotURL,
		&check.ContentHash,
		&check.ChangeDetected,
		&check.ChangeType,
		&check.ErrorMessage,
		&check.DurationMs,
		&check.ScreenshotHash,
		&check.VisionChangeSummary,
		&check.CheckedAt,
	)
}

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

	q := `INSERT INTO checks (id, page_id, section_id, parent_check_id, status, screenshot_url, html_snapshot_url, content_hash, change_detected, change_type, error_message, duration_ms, screenshot_hash, vision_change_summary, checked_at)
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err := r.db.ExecContext(ctx, q,
		check.ID,
		check.PageID,
		check.SectionID,
		check.ParentCheckID,
		check.Status,
		check.ScreenshotURL,
		check.HTMLSnapshotURL,
		check.ContentHash,
		check.ChangeDetected,
		check.ChangeType,
		check.ErrorMessage,
		check.DurationMs,
		check.ScreenshotHash,
		check.VisionChangeSummary,
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
	q := `SELECT ` + checkSelectColumns + ` FROM checks WHERE id = $1`

	if err := scanCheck(r.db.QueryRowContext(ctx, q, id), &check); err != nil {
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
		duration_ms = $8,
		screenshot_hash = $9,
		vision_change_summary = $10,
		section_id = $11,
		parent_check_id = $12
		WHERE id = $13`

	_, err := r.db.ExecContext(ctx, q,
		check.Status,
		check.ScreenshotURL,
		check.HTMLSnapshotURL,
		check.ContentHash,
		check.ChangeDetected,
		check.ChangeType,
		check.ErrorMessage,
		check.DurationMs,
		check.ScreenshotHash,
		check.VisionChangeSummary,
		check.SectionID,
		check.ParentCheckID,
		check.ID,
	)
	return err
}

// ListByPage retrieves parent checks for a page (excludes section-level checks).
func (r *CheckPostgresRepository) ListByPage(ctx context.Context, pageID uuid.UUID) ([]*entities.Check, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	q := `SELECT ` + checkSelectColumns + ` FROM checks WHERE page_id = $1 AND section_id IS NULL ORDER BY checked_at DESC`

	rows, err := r.db.QueryContext(ctx, q, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []*entities.Check
	for rows.Next() {
		var check entities.Check
		if err := scanCheck(rows, &check); err != nil {
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
	q := `SELECT ` + checkSelectColumns + ` FROM checks WHERE page_id = $1 AND id != $2 AND status = 'success' AND section_id IS NULL ORDER BY checked_at DESC LIMIT 1`

	if err := scanCheck(r.db.QueryRowContext(ctx, q, pageID, excludeCheckID), &check); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &check, nil
}

// ListByPageAndSection retrieves checks for a page filtered by section.
// If sectionID is nil, returns only full-page checks (section_id IS NULL).
func (r *CheckPostgresRepository) ListByPageAndSection(ctx context.Context, pageID uuid.UUID, sectionID *uuid.UUID) ([]*entities.Check, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	var q string
	var rows *sql.Rows
	var err error

	if sectionID == nil {
		q = `SELECT ` + checkSelectColumns + ` FROM checks WHERE page_id = $1 AND section_id IS NULL ORDER BY checked_at DESC`
		rows, err = r.db.QueryContext(ctx, q, pageID)
	} else {
		q = `SELECT ` + checkSelectColumns + ` FROM checks WHERE page_id = $1 AND section_id = $2 ORDER BY checked_at DESC`
		rows, err = r.db.QueryContext(ctx, q, pageID, *sectionID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []*entities.Check
	for rows.Next() {
		var check entities.Check
		if err := scanCheck(rows, &check); err != nil {
			return nil, err
		}
		checks = append(checks, &check)
	}
	return checks, nil
}

// GetPreviousSuccessfulBySection retrieves the most recent successful check
// for the same page+section, excluding the given check ID.
func (r *CheckPostgresRepository) GetPreviousSuccessfulBySection(ctx context.Context, pageID uuid.UUID, sectionID *uuid.UUID, excludeCheckID uuid.UUID) (*entities.Check, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	var check entities.Check
	var q string
	var row *sql.Row

	if sectionID == nil {
		q = `SELECT ` + checkSelectColumns + ` FROM checks WHERE page_id = $1 AND section_id IS NULL AND id != $2 AND status = 'success' ORDER BY checked_at DESC LIMIT 1`
		row = r.db.QueryRowContext(ctx, q, pageID, excludeCheckID)
	} else {
		q = `SELECT ` + checkSelectColumns + ` FROM checks WHERE page_id = $1 AND section_id = $2 AND id != $3 AND status = 'success' ORDER BY checked_at DESC LIMIT 1`
		row = r.db.QueryRowContext(ctx, q, pageID, *sectionID, excludeCheckID)
	}

	if err := scanCheck(row, &check); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &check, nil
}

// GetLatestByPage retrieves the latest parent check for a page (excludes section-level checks).
func (r *CheckPostgresRepository) GetLatestByPage(ctx context.Context, pageID uuid.UUID) (*entities.Check, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	var check entities.Check
	q := `SELECT ` + checkSelectColumns + ` FROM checks WHERE page_id = $1 AND section_id IS NULL ORDER BY checked_at DESC LIMIT 1`

	if err := scanCheck(r.db.QueryRowContext(ctx, q, pageID), &check); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &check, nil
}

// ListByParentCheckID retrieves all section checks that belong to a parent check.
func (r *CheckPostgresRepository) ListByParentCheckID(ctx context.Context, parentCheckID uuid.UUID) ([]*entities.Check, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	q := `SELECT ` + checkSelectColumns + ` FROM checks WHERE parent_check_id = $1 ORDER BY checked_at ASC`

	rows, err := r.db.QueryContext(ctx, q, parentCheckID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []*entities.Check
	for rows.Next() {
		var check entities.Check
		if err := scanCheck(rows, &check); err != nil {
			return nil, err
		}
		checks = append(checks, &check)
	}
	return checks, nil
}

// ListSectionChecksByPage returns all section checks for a page (section_id IS NOT NULL).
func (r *CheckPostgresRepository) ListSectionChecksByPage(ctx context.Context, pageID uuid.UUID) ([]*entities.Check, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	q := `SELECT ` + checkSelectColumns + ` FROM checks WHERE page_id = $1 AND section_id IS NOT NULL ORDER BY checked_at DESC`

	rows, err := r.db.QueryContext(ctx, q, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var checks []*entities.Check
	for rows.Next() {
		var check entities.Check
		if err := scanCheck(rows, &check); err != nil {
			return nil, err
		}
		checks = append(checks, &check)
	}
	return checks, nil
}
