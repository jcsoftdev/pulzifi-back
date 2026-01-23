package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
)

type InsightPostgresRepository struct {
	db     *sql.DB
	tenant string
}

func NewInsightPostgresRepository(db *sql.DB, tenant string) repositories.InsightRepository {
	return &InsightPostgresRepository{db: db, tenant: tenant}
}

func (r *InsightPostgresRepository) Create(ctx context.Context, insight *entities.Insight) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}
	q := `INSERT INTO insights (id, page_id, check_id, insight_type, title, content, metadata, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, q, insight.ID, insight.PageID, insight.CheckID, insight.InsightType, insight.Title, insight.Content, insight.Metadata, insight.CreatedAt)
	return err
}

func (r *InsightPostgresRepository) ListByPageID(ctx context.Context, pageID uuid.UUID) ([]*entities.Insight, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}
	q := `SELECT id, page_id, check_id, insight_type, title, content, metadata, created_at FROM insights WHERE page_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, q, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var insights []*entities.Insight
	for rows.Next() {
		var i entities.Insight
		if err := rows.Scan(&i.ID, &i.PageID, &i.CheckID, &i.InsightType, &i.Title, &i.Content, &i.Metadata, &i.CreatedAt); err != nil {
			return nil, err
		}
		insights = append(insights, &i)
	}
	return insights, nil
}

func (r *InsightPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Insight, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}
	q := `SELECT id, page_id, check_id, insight_type, title, content, metadata, created_at FROM insights WHERE id = $1 AND deleted_at IS NULL`
	var i entities.Insight
	err := r.db.QueryRowContext(ctx, q, id).Scan(&i.ID, &i.PageID, &i.CheckID, &i.InsightType, &i.Title, &i.Content, &i.Metadata, &i.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &i, nil
}
