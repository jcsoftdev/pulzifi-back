package jobs

import (
	"context"
	"database/sql"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/application/retention"
	snapshotpersistence "github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/persistence"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/repositories"
	snapshotstorage "github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/storage"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// SnapshotRetentionRunner runs the snapshot retention sweep on a configurable interval.
type SnapshotRetentionRunner struct {
	job      *retention.RetentionJob
	db       *sql.DB
	interval time.Duration
}

// NewSnapshotRetentionRunner wires up a SnapshotRetentionRunner from config.
// It resolves the object storage provider and creates the per-tenant repo factory.
func NewSnapshotRetentionRunner(db *sql.DB, cfg *config.Config) (*SnapshotRetentionRunner, error) {
	objectStorage, err := snapshotstorage.NewObjectStorage(cfg)
	if err != nil {
		return nil, err
	}

	deps := retention.RetentionJobDeps{
		RepoFactory: func(schemaName string) repositories.RetentionRepository {
			return snapshotpersistence.NewRetentionPostgresRepository(db, schemaName)
		},
		Storage:              objectStorage,
		DefaultRetentionDays: cfg.DefaultRetentionDays,
		// Tenants list is refreshed before each run inside Run().
	}

	return &SnapshotRetentionRunner{
		job:      retention.NewRetentionJob(deps),
		db:       db,
		interval: cfg.RetentionJobInterval,
	}, nil
}

// Run loops on the configured interval until ctx is cancelled.
// It runs one tick immediately so an operator can verify the wiring on startup.
func (r *SnapshotRetentionRunner) Run(ctx context.Context) error {
	logger.Info("Starting snapshot retention runner", zap.Duration("interval", r.interval))
	if err := r.runOnce(ctx); err != nil {
		logger.Warn("snapshot retention: first tick errored", zap.Error(err))
	}
	t := time.NewTicker(r.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Info("Snapshot retention runner stopping")
			return ctx.Err()
		case <-t.C:
			if err := r.runOnce(ctx); err != nil {
				logger.Warn("snapshot retention: tick errored", zap.Error(err))
			}
		}
	}
}

// runOnce refreshes the tenant list and runs one retention sweep.
func (r *SnapshotRetentionRunner) runOnce(ctx context.Context) error {
	tenants, err := r.listTenants(ctx)
	if err != nil {
		return err
	}
	r.job.SetTenants(tenants)
	return r.job.RunOnce(ctx)
}

// listTenants queries public.organizations for all active tenant schema names.
func (r *SnapshotRetentionRunner) listTenants(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT schema_name FROM public.organizations WHERE deleted_at IS NULL AND schema_name <> ''`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tenants []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		tenants = append(tenants, s)
	}
	return tenants, rows.Err()
}
