package persistence

import (
	"database/sql"

	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/orchestrator"
)

type PostgresRepositoryFactory struct {
	db *sql.DB
}

func NewPostgresRepositoryFactory(db *sql.DB) *PostgresRepositoryFactory {
	return &PostgresRepositoryFactory{db: db}
}

func (f *PostgresRepositoryFactory) GetCheckRepository(tenant string) orchestrator.CheckRepository {
	return NewCheckPostgresRepository(f.db, tenant)
}

func (f *PostgresRepositoryFactory) GetPageRepository(tenant string) orchestrator.PageRepository {
	return NewMonitoringPagePostgresRepository(f.db, tenant)
}

func (f *PostgresRepositoryFactory) GetUsageRepository(tenant string) orchestrator.UsageRepository {
	return NewUsagePostgresRepository(f.db, tenant)
}
