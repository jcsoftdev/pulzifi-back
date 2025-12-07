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

func (r *PagePostgresRepository) Create(ctx context.Context, page *entities.Page) error {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return err
	}
	q := `INSERT INTO pages (id, workspace_id, name, url, check_count, created_by, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, q, page.ID, page.WorkspaceID, page.Name, page.URL, page.CheckCount, page.CreatedBy, page.CreatedAt, page.UpdatedAt)
	return err
}

func (r *PagePostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Page, error) {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return nil, err
	}
	var p entities.Page
	var del sql.NullTime
	q := `SELECT id, workspace_id, name, url, check_count, created_by, created_at, updated_at, deleted_at FROM pages WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.QueryRowContext(ctx, q, id).Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.URL, &p.CheckCount, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt, &del)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if del.Valid {
		p.DeletedAt = &del.Time
	}
	return &p, nil
}

func (r *PagePostgresRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entities.Page, error) {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return nil, err
	}
	q := `SELECT id, workspace_id, name, url, check_count, created_by, created_at, updated_at, deleted_at FROM pages WHERE workspace_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, q, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pages []*entities.Page
	for rows.Next() {
		var p entities.Page
		var del sql.NullTime
		if err := rows.Scan(&p.ID, &p.WorkspaceID, &p.Name, &p.URL, &p.CheckCount, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt, &del); err != nil {
			return nil, err
		}
		if del.Valid {
			p.DeletedAt = &del.Time
		}
		pages = append(pages, &p)
	}
	return pages, nil
}

func (r *PagePostgresRepository) Update(ctx context.Context, page *entities.Page) error {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return err
	}
	page.UpdatedAt = time.Now()
	q := `UPDATE pages SET name = $1, url = $2, check_count = $3, updated_at = $4 WHERE id = $5`
	_, err := r.db.ExecContext(ctx, q, page.Name, page.URL, page.CheckCount, page.UpdatedAt, page.ID)
	return err
}

func (r *PagePostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, "SET search_path TO "+r.tenant); err != nil {
		return err
	}
	q := `UPDATE pages SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`
	_, err := r.db.ExecContext(ctx, q, time.Now(), id)
	return err
}
