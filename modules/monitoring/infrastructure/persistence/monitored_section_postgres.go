package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
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
	offsetsJSON := marshalSelectorOffsets(section.SelectorOffsets)
	rectJSON := marshalSectionRect(section.Rect)
	q := `INSERT INTO monitored_sections (id, page_id, name, css_selector, xpath_selector, selector_offsets, rect, viewport_width, sort_order, created_at, updated_at)
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q,
			section.ID, section.PageID, section.Name, section.CSSSelector, section.XPathSelector,
			string(offsetsJSON), rectJSON, section.ViewportWidth, section.SortOrder, section.CreatedAt, section.UpdatedAt,
		)
		return err
	})
}

func (r *MonitoredSectionPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.MonitoredSection, error) {
	var s entities.MonitoredSection
	q := `SELECT id, page_id, name, css_selector, xpath_selector,
	             COALESCE(selector_offsets, '{"top":0,"right":0,"bottom":0,"left":0}')::text,
	             rect, COALESCE(viewport_width, 0),
	             sort_order, created_at, updated_at
	      FROM monitored_sections WHERE id = $1`

	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		if err := scanSection(tx.QueryRowContext(ctx, q, id), &s); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return sql.ErrNoRows
			}
			return err
		}
		return nil
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *MonitoredSectionPostgresRepository) ListByPageID(ctx context.Context, pageID uuid.UUID) ([]*entities.MonitoredSection, error) {
	q := `SELECT id, page_id, name, css_selector, xpath_selector,
	             COALESCE(selector_offsets, '{"top":0,"right":0,"bottom":0,"left":0}')::text,
	             rect, COALESCE(viewport_width, 0),
	             sort_order, created_at, updated_at
	      FROM monitored_sections WHERE page_id = $1 ORDER BY sort_order ASC, created_at ASC`

	var sections []*entities.MonitoredSection
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, q, pageID)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var s entities.MonitoredSection
			if err := scanSection(rows, &s); err != nil {
				return err
			}
			sections = append(sections, &s)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return sections, nil
}

func (r *MonitoredSectionPostgresRepository) Update(ctx context.Context, section *entities.MonitoredSection) error {
	section.UpdatedAt = time.Now()
	offsetsJSON := marshalSelectorOffsets(section.SelectorOffsets)
	rectJSON := marshalSectionRect(section.Rect)
	q := `UPDATE monitored_sections
	      SET name = $1, css_selector = $2, xpath_selector = $3, selector_offsets = $4,
	          rect = $5, viewport_width = $6, sort_order = $7, updated_at = $8
	      WHERE id = $9`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q,
			section.Name, section.CSSSelector, section.XPathSelector, string(offsetsJSON),
			rectJSON, section.ViewportWidth, section.SortOrder, section.UpdatedAt, section.ID,
		)
		return err
	})
}

func (r *MonitoredSectionPostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	q := `DELETE FROM monitored_sections WHERE id = $1`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q, id)
		return err
	})
}

func (r *MonitoredSectionPostgresRepository) ReplaceAll(ctx context.Context, pageID uuid.UUID, sections []*entities.MonitoredSection) error {
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
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

		return nil
	})
}
