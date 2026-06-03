package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
)

type InsightPostgresRepository struct {
	db     *sql.DB
	tenant string
}

func NewInsightPostgresRepository(db *sql.DB, tenant string) repositories.InsightRepository {
	return &InsightPostgresRepository{db: db, tenant: tenant}
}

func (r *InsightPostgresRepository) Create(ctx context.Context, insight *entities.Insight) error {
	// Ensure metadata is valid JSON — nil/empty byte slice is not accepted by postgres json columns.
	metadata := insight.Metadata
	if len(metadata) == 0 {
		metadata = json.RawMessage("null")
	}
	q := `INSERT INTO insights (id, page_id, check_id, insight_type, title, content, metadata, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q, insight.ID, insight.PageID, insight.CheckID, insight.InsightType, insight.Title, insight.Content, metadata, insight.CreatedAt)
		return err
	})
}

func (r *InsightPostgresRepository) ListByPageID(ctx context.Context, pageID uuid.UUID) ([]*entities.Insight, error) {
	q := `SELECT id, page_id, check_id, insight_type, title, content, metadata, created_at FROM insights WHERE page_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	var insights []*entities.Insight
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, q, pageID)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var i entities.Insight
			if err := rows.Scan(&i.ID, &i.PageID, &i.CheckID, &i.InsightType, &i.Title, &i.Content, &i.Metadata, &i.CreatedAt); err != nil {
				return err
			}
			insights = append(insights, &i)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return insights, nil
}

func (r *InsightPostgresRepository) ListByCheckID(ctx context.Context, checkID uuid.UUID) ([]*entities.Insight, error) {
	q := `SELECT id, page_id, check_id, insight_type, title, content, metadata, created_at FROM insights WHERE check_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	var insights []*entities.Insight
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		rows, err := tx.QueryContext(ctx, q, checkID)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var i entities.Insight
			if err := rows.Scan(&i.ID, &i.PageID, &i.CheckID, &i.InsightType, &i.Title, &i.Content, &i.Metadata, &i.CreatedAt); err != nil {
				return err
			}
			insights = append(insights, &i)
		}
		return rows.Err()
	})
	if err != nil {
		return nil, err
	}
	return insights, nil
}

func (r *InsightPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Insight, error) {
	q := `SELECT id, page_id, check_id, insight_type, title, content, metadata, created_at FROM insights WHERE id = $1 AND deleted_at IS NULL`
	var result *entities.Insight
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		var i entities.Insight
		err := tx.QueryRowContext(ctx, q, id).Scan(&i.ID, &i.PageID, &i.CheckID, &i.InsightType, &i.Title, &i.Content, &i.Metadata, &i.CreatedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return err
		}
		result = &i
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}
