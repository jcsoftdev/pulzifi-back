// Package insightwiring provides cross-module adapters for the insight module.
// It bridges monitoring infrastructure to insight's domain interfaces,
// keeping insight free of cross-module imports.
package insightwiring

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	insightservices "github.com/jcsoftdev/pulzifi-back/modules/insight/domain/services"
	monpersistence "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/persistence"
)

// --- PageConfigReaderAdapter ---

// PageConfigReaderAdapter adapts MonitoringConfigPostgresRepository to
// insight's PageConfigReader interface.
type PageConfigReaderAdapter struct {
	db     *sql.DB
	tenant string
}

// NewPageConfigReaderAdapter creates an adapter for the given tenant.
func NewPageConfigReaderAdapter(db *sql.DB, tenant string) insightservices.PageConfigReader {
	return &PageConfigReaderAdapter{db: db, tenant: tenant}
}

func (a *PageConfigReaderAdapter) GetInsightConfig(ctx context.Context, pageID uuid.UUID) (*insightservices.PageInsightConfig, error) {
	repo := monpersistence.NewMonitoringConfigPostgresRepository(a.db, a.tenant)
	cfg, err := repo.GetByPageID(ctx, pageID)
	if err != nil || cfg == nil {
		return nil, err
	}
	return &insightservices.PageInsightConfig{
		EnabledInsightTypes: cfg.EnabledInsightTypes,
	}, nil
}

func (a *PageConfigReaderAdapter) GetPageURL(ctx context.Context, pageID uuid.UUID) (string, error) {
	repo := monpersistence.NewMonitoringConfigPostgresRepository(a.db, a.tenant)
	return repo.GetPageURL(ctx, pageID)
}

// --- CheckReaderAdapter ---

// CheckReaderAdapter adapts CheckPostgresRepository to insight's CheckReader interface.
type CheckReaderAdapter struct {
	db     *sql.DB
	tenant string
}

// NewCheckReaderAdapter creates an adapter for the given tenant.
func NewCheckReaderAdapter(db *sql.DB, tenant string) insightservices.CheckReader {
	return &CheckReaderAdapter{db: db, tenant: tenant}
}

func (a *CheckReaderAdapter) GetByID(ctx context.Context, checkID uuid.UUID) (*insightservices.CheckSummary, error) {
	repo := monpersistence.NewCheckPostgresRepository(a.db, a.tenant)
	check, err := repo.GetByID(ctx, checkID)
	if err != nil || check == nil {
		return nil, err
	}
	return &insightservices.CheckSummary{
		ID:              check.ID,
		Status:          check.Status,
		HTMLSnapshotURL: check.HTMLSnapshotURL,
	}, nil
}

func (a *CheckReaderAdapter) ListByPage(ctx context.Context, pageID uuid.UUID) ([]*insightservices.CheckSummary, error) {
	repo := monpersistence.NewCheckPostgresRepository(a.db, a.tenant)
	checks, err := repo.ListByPage(ctx, pageID)
	if err != nil {
		return nil, err
	}
	result := make([]*insightservices.CheckSummary, len(checks))
	for i, c := range checks {
		result[i] = &insightservices.CheckSummary{
			ID:              c.ID,
			Status:          c.Status,
			HTMLSnapshotURL: c.HTMLSnapshotURL,
		}
	}
	return result, nil
}
