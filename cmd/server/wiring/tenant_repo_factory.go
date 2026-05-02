package wiring

import (
	"database/sql"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/persistence"
)

// TenantRepoFactory satisfies dispatch_event.RepoFactory — returns tenant-scoped repos on demand.
type TenantRepoFactory struct{ db *sql.DB }

// NewTenantRepoFactory constructs a TenantRepoFactory backed by the given DB pool.
func NewTenantRepoFactory(db *sql.DB) *TenantRepoFactory {
	return &TenantRepoFactory{db: db}
}

// DestinationRepoForTenant returns a tenant-scoped DestinationRepository.
func (f *TenantRepoFactory) DestinationRepoForTenant(tenant string) repositories.DestinationRepository {
	return persistence.NewDestinationPostgresRepository(f.db, tenant)
}

// DeliveryRepoForTenant returns a tenant-scoped DeliveryRepository.
func (f *TenantRepoFactory) DeliveryRepoForTenant(tenant string) repositories.DeliveryRepository {
	return persistence.NewDeliveryPostgresRepository(f.db, tenant)
}
