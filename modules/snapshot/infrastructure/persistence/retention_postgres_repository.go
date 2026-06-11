package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
)

// RetentionPostgresRepository implements repositories.RetentionRepository for
// a specific tenant schema using a shared *sql.DB pool.
type RetentionPostgresRepository struct {
	db         *sql.DB
	schemaName string
}

// NewRetentionPostgresRepository creates a RetentionRepository scoped to one tenant.
func NewRetentionPostgresRepository(db *sql.DB, schemaName string) repositories.RetentionRepository {
	return &RetentionPostgresRepository{db: db, schemaName: schemaName}
}

// GetStoragePeriodDays reads the tenant's storage_period_days from usage_tracking.
// Returns 0 when no active row exists so the caller can apply the default.
func (r *RetentionPostgresRepository) GetStoragePeriodDays(ctx context.Context) (int, error) {
	var days int
	err := database.WithTenant(ctx, r.db, r.schemaName, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, `
			SELECT COALESCE(storage_period_days, 0)
			  FROM usage_tracking
			 ORDER BY created_at DESC
			 LIMIT 1
		`).Scan(&days)
		if err == sql.ErrNoRows {
			return nil
		}
		if err != nil {
			return fmt.Errorf("retention: get storage_period_days for %s: %w", r.schemaName, err)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return days, nil
}

// ListExpiredChecks returns checks older than cutoff that still have at least
// one storage URL to clean up. Only top-level (non-section) checks are targeted
// to avoid double-counting; section checks share the same bucket path.
//
// Predecessor protection: a check that is the direct predecessor (most recent
// previous check on the same page) of a change-detected check still within the
// retention window is excluded. Without this guard the "previous snapshot" shown
// in the before/after view would be deleted before the change itself expires,
// leaving users with an empty comparison view.
func (r *RetentionPostgresRepository) ListExpiredChecks(ctx context.Context, cutoff time.Time) ([]repositories.ExpiredCheck, error) {
	var out []repositories.ExpiredCheck
	err := database.WithTenant(ctx, r.db, r.schemaName, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, `
			SELECT id,
			       COALESCE(screenshot_url, ''),
			       COALESCE(html_snapshot_url, ''),
			       COALESCE(diff_image_url, '')
			  FROM checks
			 WHERE checked_at < $1
			   AND section_id IS NULL
			   AND (
			         screenshot_url    IS NOT NULL
			      OR html_snapshot_url IS NOT NULL
			      OR diff_image_url    IS NOT NULL
			   )
			   AND id NOT IN (
			         SELECT prev.id
			           FROM checks AS change_check
			           JOIN LATERAL (
			                  SELECT id
			                    FROM checks AS p
			                   WHERE p.page_id    = change_check.page_id
			                     AND p.section_id IS NULL
			                     AND p.checked_at < change_check.checked_at
			                   ORDER BY p.checked_at DESC
			                   LIMIT 1
			                ) AS prev ON TRUE
			          WHERE change_check.change_detected = TRUE
			            AND change_check.checked_at      >= $1
			            AND change_check.section_id      IS NULL
			       )
		`, cutoff)
		if err != nil {
			return fmt.Errorf("retention: list expired checks for %s: %w", r.schemaName, err)
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var c repositories.ExpiredCheck
			if err := rows.Scan(&c.ID, &c.ScreenshotURL, &c.HTMLSnapshotURL, &c.DiffImageURL); err != nil {
				return fmt.Errorf("retention: scan check row: %w", err)
			}
			out = append(out, c)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("retention: iterate checks: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NullifyStorageURLs clears the three storage URL columns for the given check IDs
// in a single bulk update using an ANY($1) clause.
func (r *RetentionPostgresRepository) NullifyStorageURLs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	return database.WithTenant(ctx, r.db, r.schemaName, func(tx *sql.Tx) error {
		// Build a PostgreSQL array literal from validated UUID values.
		// UUIDs are format-validated by the uuid.UUID type so this is injection-safe.
		arrayLiteral := buildUUIDArrayLiteral(ids)
		_, err := tx.ExecContext(ctx, `
			UPDATE checks
			   SET screenshot_url    = NULL,
			       html_snapshot_url = NULL,
			       diff_image_url    = NULL
			 WHERE id = ANY(`+arrayLiteral+`)
		`)
		if err != nil {
			return fmt.Errorf("retention: nullify storage URLs for %s: %w", r.schemaName, err)
		}
		return nil
	})
}

// buildUUIDArrayLiteral returns a PostgreSQL UUID array literal, e.g.
// ARRAY['uuid1','uuid2']::uuid[]
func buildUUIDArrayLiteral(ids []uuid.UUID) string {
	if len(ids) == 0 {
		return "ARRAY[]::uuid[]"
	}
	lit := "ARRAY["
	for i, id := range ids {
		if i > 0 {
			lit += ","
		}
		lit += "'" + id.String() + "'"
	}
	lit += "]::uuid[]"
	return lit
}
