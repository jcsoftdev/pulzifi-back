package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
)

type MonitoredSectionPostgresRepository struct {
	db     *sql.DB
	tenant string
}

func NewMonitoredSectionPostgresRepository(db *sql.DB, tenant string) *MonitoredSectionPostgresRepository {
	return &MonitoredSectionPostgresRepository{db: db, tenant: tenant}
}

func marshalSectionRect(r *entities.SectionRect) []byte {
	if r == nil {
		return nil
	}
	b, _ := json.Marshal(r)
	return b
}

func scanSection(row interface{ Scan(...interface{}) error }, s *entities.MonitoredSection) error {
	var offsetsRaw []byte
	var rectRaw []byte
	err := row.Scan(
		&s.ID, &s.PageID, &s.Name, &s.CSSSelector, &s.XPathSelector,
		&offsetsRaw, &rectRaw, &s.ViewportWidth, &s.SortOrder, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if len(offsetsRaw) > 0 {
		var offsets entities.SelectorOffsets
		if json.Unmarshal(offsetsRaw, &offsets) == nil {
			s.SelectorOffsets = &offsets
		}
	}
	if len(rectRaw) > 0 {
		var rect entities.SectionRect
		if json.Unmarshal(rectRaw, &rect) == nil {
			s.Rect = &rect
		}
	}
	return nil
}

func (r *MonitoredSectionPostgresRepository) Create(ctx context.Context, section *entities.MonitoredSection) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}
	offsetsJSON := marshalSelectorOffsets(section.SelectorOffsets)
	rectJSON := marshalSectionRect(section.Rect)
	q := `INSERT INTO monitored_sections (id, page_id, name, css_selector, xpath_selector, selector_offsets, rect, viewport_width, sort_order, created_at, updated_at)
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.db.ExecContext(ctx, q,
		section.ID, section.PageID, section.Name, section.CSSSelector, section.XPathSelector,
		string(offsetsJSON), rectJSON, section.ViewportWidth, section.SortOrder, section.CreatedAt, section.UpdatedAt,
	)
	return err
}

func (r *MonitoredSectionPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.MonitoredSection, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}
	var s entities.MonitoredSection
	q := `SELECT id, page_id, name, css_selector, xpath_selector,
	             COALESCE(selector_offsets, '{"top":0,"right":0,"bottom":0,"left":0}')::text,
	             rect, COALESCE(viewport_width, 0),
	             sort_order, created_at, updated_at
	      FROM monitored_sections WHERE id = $1`
	if err := scanSection(r.db.QueryRowContext(ctx, q, id), &s); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *MonitoredSectionPostgresRepository) ListByPageID(ctx context.Context, pageID uuid.UUID) ([]*entities.MonitoredSection, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}
	q := `SELECT id, page_id, name, css_selector, xpath_selector,
	             COALESCE(selector_offsets, '{"top":0,"right":0,"bottom":0,"left":0}')::text,
	             rect, COALESCE(viewport_width, 0),
	             sort_order, created_at, updated_at
	      FROM monitored_sections WHERE page_id = $1 ORDER BY sort_order ASC, created_at ASC`
	rows, err := r.db.QueryContext(ctx, q, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sections []*entities.MonitoredSection
	for rows.Next() {
		var s entities.MonitoredSection
		if err := scanSection(rows, &s); err != nil {
			return nil, err
		}
		sections = append(sections, &s)
	}
	return sections, nil
}

func (r *MonitoredSectionPostgresRepository) Update(ctx context.Context, section *entities.MonitoredSection) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}
	section.UpdatedAt = time.Now()
	offsetsJSON := marshalSelectorOffsets(section.SelectorOffsets)
	rectJSON := marshalSectionRect(section.Rect)
	q := `UPDATE monitored_sections
	      SET name = $1, css_selector = $2, xpath_selector = $3, selector_offsets = $4,
	          rect = $5, viewport_width = $6, sort_order = $7, updated_at = $8
	      WHERE id = $9`
	_, err := r.db.ExecContext(ctx, q,
		section.Name, section.CSSSelector, section.XPathSelector, string(offsetsJSON),
		rectJSON, section.ViewportWidth, section.SortOrder, section.UpdatedAt, section.ID,
	)
	return err
}

func (r *MonitoredSectionPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}
	q := `DELETE FROM monitored_sections WHERE id = $1`
	_, err := r.db.ExecContext(ctx, q, id)
	return err
}

func (r *MonitoredSectionPostgresRepository) ReplaceAll(ctx context.Context, pageID uuid.UUID, sections []*entities.MonitoredSection) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing sections for this page
	if _, err := tx.ExecContext(ctx, `DELETE FROM monitored_sections WHERE page_id = $1`, pageID); err != nil {
		return err
	}

	// Insert new sections
	q := `INSERT INTO monitored_sections (id, page_id, name, css_selector, xpath_selector, selector_offsets, rect, viewport_width, sort_order, created_at, updated_at)
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	for _, s := range sections {
		offsetsJSON := marshalSelectorOffsets(s.SelectorOffsets)
		rectJSON := marshalSectionRect(s.Rect)
		if _, err := tx.ExecContext(ctx, q,
			s.ID, pageID, s.Name, s.CSSSelector, s.XPathSelector,
			string(offsetsJSON), rectJSON, s.ViewportWidth, s.SortOrder, s.CreatedAt, s.UpdatedAt,
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}
