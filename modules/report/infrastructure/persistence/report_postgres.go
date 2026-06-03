package persistence

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
)

// dbContent wraps entities.Content to provide sql.Scanner and driver.Valuer
// without polluting the domain layer with infrastructure imports.
type dbContent entities.Content

func (c dbContent) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

func (c *dbContent) Scan(value interface{}) error {
	if value == nil {
		*c = make(dbContent)
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(b, c)
}

type ReportPostgresRepository struct {
	db     *sql.DB
	tenant string
}

func NewReportPostgresRepository(db *sql.DB, tenant string) *ReportPostgresRepository {
	return &ReportPostgresRepository{db: db, tenant: tenant}
}

func (r *ReportPostgresRepository) Create(ctx context.Context, report *entities.Report) error {
	q := `INSERT INTO reports (id, page_id, title, report_date, content, pdf_url, created_by, created_at)
	      VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	return database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, q,
			report.ID, report.PageID, report.Title, report.ReportDate,
			dbContent(report.Content), report.PDFURL, report.CreatedBy, report.CreatedAt,
		)
		return err
	})
}

func (r *ReportPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Report, error) {
	var report entities.Report
	var pdfURL sql.NullString
	var content dbContent
	q := `SELECT id, page_id, title, report_date, content, pdf_url, created_by, created_at
	      FROM reports WHERE id = $1 AND deleted_at IS NULL`

	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, q, id).Scan(
			&report.ID, &report.PageID, &report.Title, &report.ReportDate,
			&content, &pdfURL, &report.CreatedBy, &report.CreatedAt,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return sql.ErrNoRows
			}
			return err
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	report.Content = entities.Content(content)
	report.PDFURL = pdfURL.String
	return &report, nil
}

func (r *ReportPostgresRepository) ListByPage(ctx context.Context, pageID uuid.UUID) ([]*entities.Report, error) {
	q := `SELECT id, page_id, title, report_date, content, pdf_url, created_by, created_at
	      FROM reports WHERE page_id = $1 AND deleted_at IS NULL ORDER BY report_date DESC`
	var reports []*entities.Report
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		var err error
		reports, err = scanReportRows(ctx, tx, q, pageID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func (r *ReportPostgresRepository) List(ctx context.Context) ([]*entities.Report, error) {
	q := `SELECT id, page_id, title, report_date, content, pdf_url, created_by, created_at
	      FROM reports WHERE deleted_at IS NULL ORDER BY report_date DESC`
	var reports []*entities.Report
	err := database.WithTenant(ctx, r.db, r.tenant, func(tx *sql.Tx) error {
		var err error
		reports, err = scanReportRows(ctx, tx, q)
		return err
	})
	if err != nil {
		return nil, err
	}
	return reports, nil
}

// scanReportRows executes q against tx and scans the result into a slice of Report.
// It replaces the old r.db-based scanRows helper so all scanning stays inside the
// WithTenant transaction boundary.
func scanReportRows(ctx context.Context, tx *sql.Tx, q string, args ...interface{}) ([]*entities.Report, error) {
	rows, err := tx.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var reports []*entities.Report
	for rows.Next() {
		var report entities.Report
		var pdfURL sql.NullString
		var content dbContent
		if err := rows.Scan(
			&report.ID, &report.PageID, &report.Title, &report.ReportDate,
			&content, &pdfURL, &report.CreatedBy, &report.CreatedAt,
		); err != nil {
			return nil, err
		}
		report.Content = entities.Content(content)
		report.PDFURL = pdfURL.String
		reports = append(reports, &report)
	}
	return reports, rows.Err()
}
