// Package snapshotwiring provides cross-module adapters for the snapshot module.
// It bridges snapshot's port interfaces to concrete implementations from
// monitoring, alert, email, and insight infrastructure without creating
// cross-module imports inside the snapshot module itself.
package snapshotwiring

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	monPersistence "github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/persistence"
	snapEntities "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/entities"
	snapServices "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
)

// checkRepoAdapter adapts monitoring's CheckPostgresRepository to
// snapshot's CheckRepository port.
type checkRepoAdapter struct {
	inner *monPersistence.CheckPostgresRepository
}

// NewCheckRepoFactory returns a factory function that produces a CheckRepository
// for the given tenant schema. This is the value passed as SnapshotWorkerDeps.CheckRepoFactory.
func NewCheckRepoFactory(db *sql.DB) func(schemaName string) snapServices.CheckRepository {
	return func(schemaName string) snapServices.CheckRepository {
		return &checkRepoAdapter{
			inner: monPersistence.NewCheckPostgresRepository(db, schemaName),
		}
	}
}

func (a *checkRepoAdapter) Create(ctx context.Context, check *snapEntities.Check) error {
	return a.inner.Create(ctx, toMonCheck(check))
}

func (a *checkRepoAdapter) GetByID(ctx context.Context, id uuid.UUID) (*snapEntities.Check, error) {
	mc, err := a.inner.GetByID(ctx, id)
	if err != nil || mc == nil {
		return nil, err
	}
	return fromMonCheck(mc), nil
}

func (a *checkRepoAdapter) Update(ctx context.Context, check *snapEntities.Check) error {
	return a.inner.Update(ctx, toMonCheck(check))
}

func (a *checkRepoAdapter) GetPreviousSuccessfulByPage(ctx context.Context, pageID, excludeID uuid.UUID) (*snapEntities.Check, error) {
	mc, err := a.inner.GetPreviousSuccessfulByPage(ctx, pageID, excludeID)
	if err != nil || mc == nil {
		return nil, err
	}
	return fromMonCheck(mc), nil
}

func (a *checkRepoAdapter) GetPreviousSuccessfulBySection(ctx context.Context, pageID uuid.UUID, sectionID *uuid.UUID, excludeID uuid.UUID) (*snapEntities.Check, error) {
	mc, err := a.inner.GetPreviousSuccessfulBySection(ctx, pageID, sectionID, excludeID)
	if err != nil || mc == nil {
		return nil, err
	}
	return fromMonCheck(mc), nil
}
