package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/middleware"
)

type ReportPostgresRepository struct {
	db     *sql.DB
	tenant string
}

func NewReportPostgresRepository(db *sql.DB, tenant string) *ReportPostgresRepository {
	return &ReportPostgresRepository{db: db, tenant: tenant}
}

func (r *ReportPostgresRepository) Create(ctx context.Context, report *entities.Report) error {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return err
	}

	q := `INSERT INTO reports (id, page_id, title, report_date, content, pdf_url, created_by, created_at)
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.ExecContext(ctx, q,
		report.ID, report.PageID, report.Title, report.ReportDate,
		report.Content, report.PDFURL, report.CreatedBy, report.CreatedAt,
	)
	return err
}

func (r *ReportPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Report, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	var report entities.Report
	var pdfURL sql.NullString
	q := `SELECT id, page_id, title, report_date, content, pdf_url, created_by, created_at
	      FROM reports WHERE id = $1 AND deleted_at IS NULL`
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&report.ID, &report.PageID, &report.Title, &report.ReportDate,
		&report.Content, &pdfURL, &report.CreatedBy, &report.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	report.PDFURL = pdfURL.String
	return &report, nil
}

func (r *ReportPostgresRepository) ListByPage(ctx context.Context, pageID uuid.UUID) ([]*entities.Report, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	q := `SELECT id, page_id, title, report_date, content, pdf_url, created_by, created_at
	      FROM reports WHERE page_id = $1 AND deleted_at IS NULL ORDER BY report_date DESC`
	return r.scanRows(ctx, q, pageID)
}

func (r *ReportPostgresRepository) List(ctx context.Context) ([]*entities.Report, error) {
	if _, err := r.db.ExecContext(ctx, middleware.GetSetSearchPathSQL(r.tenant)); err != nil {
		return nil, err
	}

	q := `SELECT id, page_id, title, report_date, content, pdf_url, created_by, created_at
	      FROM reports WHERE deleted_at IS NULL ORDER BY report_date DESC`
	return r.scanRows(ctx, q)
}

func (r *ReportPostgresRepository) scanRows(ctx context.Context, q string, args ...interface{}) ([]*entities.Report, error) {
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*entities.Report
	for rows.Next() {
		var report entities.Report
		var pdfURL sql.NullString
		if err := rows.Scan(
			&report.ID, &report.PageID, &report.Title, &report.ReportDate,
			&report.Content, &pdfURL, &report.CreatedBy, &report.CreatedAt,
		); err != nil {
			return nil, err
		}
		report.PDFURL = pdfURL.String
		reports = append(reports, &report)
	}
	return reports, rows.Err()
}
