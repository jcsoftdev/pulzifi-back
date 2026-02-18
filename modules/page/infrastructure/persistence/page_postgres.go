package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/entities"
)

type PagePostgresRepository struct {
	db     *sql.DB
	tenant string
}

func NewPagePostgresRepository(db *sql.DB, tenant string) *PagePostgresRepository {
	return &PagePostgresRepository{db: db, tenant: tenant}
}

// Helper to qualify table names with schema
func (r *PagePostgresRepository) table(name string) string {
	return `"` + r.tenant + `".` + name
}

func (r *PagePostgresRepository) Create(ctx context.Context, page *entities.Page) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	q := `INSERT INTO ` + r.table("pages") + ` (id, workspace_id, name, url, check_count, created_by, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	if _, err := tx.ExecContext(ctx, q, page.ID, page.WorkspaceID, page.Name, page.URL, page.CheckCount, page.CreatedBy, page.CreatedAt, page.UpdatedAt); err != nil {
		return err
	}

	if len(page.Tags) > 0 {
		insQ := `INSERT INTO ` + r.table("page_tags") + ` (page_id, tag) VALUES ($1, $2)`
		for _, tag := range page.Tags {
			if _, err := tx.ExecContext(ctx, insQ, page.ID, tag); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *PagePostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Page, error) {
	var p entities.Page
	var del sql.NullTime
	var thumbnailURL, checkFrequency sql.NullString
	var lastCheckedAt, lastChangeDetectedAt sql.NullTime

	q := `
		SELECT 
			p.id, p.workspace_id, p.name, p.url, p.thumbnail_url, 
			p.last_checked_at, p.last_change_detected_at, p.check_count, 
			p.created_by, p.created_at, p.updated_at, p.deleted_at,
			COALESCE(mc.check_frequency, 'Off') as check_frequency,
			COALESCE(
				(SELECT COUNT(*) FROM ` + r.table("checks") + ` WHERE page_id = p.id AND change_detected = true), 
				0
			) as detected_changes
		FROM ` + r.table("pages") + ` p
		LEFT JOIN ` + r.table("monitoring_configs") + ` mc ON mc.page_id = p.id AND mc.deleted_at IS NULL
		WHERE p.id = $1 AND p.deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&p.ID, &p.WorkspaceID, &p.Name, &p.URL, &thumbnailURL,
		&lastCheckedAt, &lastChangeDetectedAt, &p.CheckCount,
		&p.CreatedBy, &p.CreatedAt, &p.UpdatedAt, &del,
		&checkFrequency, &p.DetectedChanges,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if del.Valid {
		p.DeletedAt = &del.Time
	}
	if thumbnailURL.Valid {
		p.ThumbnailURL = thumbnailURL.String
	}
	if lastCheckedAt.Valid {
		p.LastCheckedAt = &lastCheckedAt.Time
	}
	if lastChangeDetectedAt.Valid {
		p.LastChangeDetectedAt = &lastChangeDetectedAt.Time
	}
	if checkFrequency.Valid {
		p.CheckFrequency = checkFrequency.String
	}

	// Get tags
	tagsQ := `SELECT tag FROM ` + r.table("page_tags") + ` WHERE page_id = $1 ORDER BY created_at`
	rows, err := r.db.QueryContext(ctx, tagsQ, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	p.Tags = []string{}
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		p.Tags = append(p.Tags, tag)
	}

	return &p, nil
}

func (r *PagePostgresRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entities.Page, error) {
	q := `
		SELECT 
			p.id, p.workspace_id, p.name, p.url, p.thumbnail_url, 
			p.last_checked_at, p.last_change_detected_at, p.check_count, 
			p.created_by, p.created_at, p.updated_at, p.deleted_at,
			COALESCE(mc.check_frequency, 'Off') as check_frequency,
			COALESCE(
				(SELECT COUNT(*) FROM ` + r.table("checks") + ` c WHERE c.page_id = p.id AND c.change_detected = true), 
				0
			) as detected_changes
		FROM ` + r.table("pages") + ` p
		LEFT JOIN ` + r.table("monitoring_configs") + ` mc ON mc.page_id = p.id AND mc.deleted_at IS NULL
		WHERE p.workspace_id = $1 AND p.deleted_at IS NULL 
		ORDER BY p.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, q, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []*entities.Page
	for rows.Next() {
		var p entities.Page
		var del sql.NullTime
		var thumbnailURL, checkFrequency sql.NullString
		var lastCheckedAt, lastChangeDetectedAt sql.NullTime

		if err := rows.Scan(
			&p.ID, &p.WorkspaceID, &p.Name, &p.URL, &thumbnailURL,
			&lastCheckedAt, &lastChangeDetectedAt, &p.CheckCount,
			&p.CreatedBy, &p.CreatedAt, &p.UpdatedAt, &del,
			&checkFrequency, &p.DetectedChanges,
		); err != nil {
			return nil, err
		}

		if del.Valid {
			p.DeletedAt = &del.Time
		}
		if thumbnailURL.Valid {
			p.ThumbnailURL = thumbnailURL.String
		}
		if lastCheckedAt.Valid {
			p.LastCheckedAt = &lastCheckedAt.Time
		}
		if lastChangeDetectedAt.Valid {
			p.LastChangeDetectedAt = &lastChangeDetectedAt.Time
		}
		if checkFrequency.Valid {
			p.CheckFrequency = checkFrequency.String
		}

		// Get tags for this page
		p.Tags = []string{}
		tagsQ := `SELECT tag FROM ` + r.table("page_tags") + ` WHERE page_id = $1 ORDER BY created_at`
		tagRows, err := r.db.QueryContext(ctx, tagsQ, p.ID)
		if err != nil {
			return nil, err
		}

		for tagRows.Next() {
			var tag string
			if err := tagRows.Scan(&tag); err != nil {
				tagRows.Close()
				return nil, err
			}
			p.Tags = append(p.Tags, tag)
		}
		tagRows.Close()

		pages = append(pages, &p)
	}

	return pages, nil
}

func (r *PagePostgresRepository) Update(ctx context.Context, page *entities.Page) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	page.UpdatedAt = time.Now()
	q := `UPDATE ` + r.table("pages") + ` SET name = $1, url = $2, check_count = $3, updated_at = $4 WHERE id = $5`
	if _, err := tx.ExecContext(ctx, q, page.Name, page.URL, page.CheckCount, page.UpdatedAt, page.ID); err != nil {
		return err
	}

	// Delete existing tags
	delQ := `DELETE FROM ` + r.table("page_tags") + ` WHERE page_id = $1`
	if _, err := tx.ExecContext(ctx, delQ, page.ID); err != nil {
		return err
	}

	// Insert new tags
	if len(page.Tags) > 0 {
		insQ := `INSERT INTO ` + r.table("page_tags") + ` (page_id, tag) VALUES ($1, $2)`
		for _, tag := range page.Tags {
			if _, err := tx.ExecContext(ctx, insQ, page.ID, tag); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *PagePostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	q := `UPDATE ` + r.table("pages") + ` SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, q, time.Now(), id)
	return err
}
