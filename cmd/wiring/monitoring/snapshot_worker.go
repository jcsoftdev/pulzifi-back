// Package monitoringwiring provides cross-module adapters for the monitoring module.
// It constructs snapshot/insight/email dependencies and injects them into the
// monitoring module via its Deps struct, keeping monitoring free of cross-module imports.
package monitoringwiring

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	snapshotwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/snapshot"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/application/workers"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/infrastructure/persistence"
	snapshotapp "github.com/jcsoftdev/pulzifi-back/modules/snapshot/application"
	snapshotextractor "github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/extractor"
	snapshotstorage "github.com/jcsoftdev/pulzifi-back/modules/snapshot/infrastructure/storage"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// SnapshotWorkerDeps groups everything NewSnapshotWorker needs.
type SnapshotWorkerDeps struct {
	DB            *sql.DB
	EventBus      *eventbus.EventBus
	EmailProvider emailservices.EmailProvider
	FrontendURL   string
	Cfg           *config.Config
}

// NewSnapshotWorker builds and configures a *snapshotapp.SnapshotWorker from
// snapshot/insight/email infrastructure via snapshot wiring adapters.
// Returns a workers.SnapshotPort so the monitoring module can consume it
// without cross-module imports.
//
// The onCheckDone callback must be set after the monitoring module's pubsub
// broker is created; call worker.SetOnCheckDone(...) from cmd/server/modules.go
// after constructing the Module.
func NewSnapshotWorker(deps SnapshotWorkerDeps) (*snapshotapp.SnapshotWorker, error) {
	cfg := deps.Cfg

	// Initialize Snapshot Infrastructure
	objectStorage, err := snapshotstorage.NewObjectStorage(cfg)
	if err != nil {
		objectStorage = nil
		logger.Error("Failed to initialize object storage client — snapshot uploads will fail", zap.Error(err))
	} else if objectStorage != nil {
		if err := objectStorage.EnsureBucket(context.Background()); err != nil {
			logger.Error("Failed to initialize object storage", zap.Error(err))
		}
	}

	extractorHTTPClient := snapshotextractor.NewHTTPClientWithKey(cfg.ExtractorURL, cfg.ExtractorAPIKey)

	// Build snapshot port adapters via the snapshotwiring package.
	extractorClient := snapshotwiring.NewExtractorClient(extractorHTTPClient)
	monitoringReader := snapshotwiring.NewMonitoringReader(deps.DB)
	checkRepoFactory := snapshotwiring.NewCheckRepoFactory(deps.DB)
	alertCreator := snapshotwiring.NewAlertCreator(deps.DB)
	insightDispatcher := snapshotwiring.NewInsightDispatcher(deps.DB, cfg)

	var notificationEmailer snapshotapp.NotificationEmailerPort
	if deps.EmailProvider != nil {
		notificationEmailer = snapshotwiring.NewNotificationEmailer(deps.EmailProvider)
	}

	workerDeps := snapshotapp.SnapshotWorkerDeps{
		ObjectStorage:       objectStorage,
		ExtractorClient:     extractorClient,
		MonitoringReader:    monitoringReader,
		CheckRepoFactory:    checkRepoFactory,
		AlertCreator:        alertCreator,
		NotificationEmailer: notificationEmailer,
		InsightDispatcher:   insightDispatcher,
		FrontendURL:         deps.FrontendURL,
	}

	snapshotWorker := snapshotapp.NewSnapshotWorker(workerDeps)
	snapshotWorker.SetPixelDiffThreshold(cfg.PixelDiffThreshold)

	if deps.EventBus != nil {
		snapshotWorker.SetMessageBus(deps.EventBus)
	}

	// Initialize Vision AI analyzer if vision model is configured.
	if visionAnalyzer := snapshotwiring.NewVisionAnalyzer(cfg); visionAnalyzer != nil {
		snapshotWorker.SetVisionAnalyzer(visionAnalyzer)
		logger.Info("Vision AI analyzer initialized", zap.String("model", cfg.OpenRouterVisionModel))
	}

	return snapshotWorker, nil
}

// NewWorkerPool builds a *workers.WorkerPool from a SnapshotPort and a fail-check
// callback. The failCheck function marks a check as "error" in the DB and notifies
// SSE subscribers; it must be provided by the monitoring module after its checkBroker
// is created.
func NewWorkerPool(
	db *sql.DB,
	snapshotPort workers.SnapshotPort,
	onCheckDone func(pageID uuid.UUID, checkJSON []byte),
	bufferSize int,
) *workers.WorkerPool {
	// Build the failCheck callback inline — it needs db and the onCheckDone hook.
	failCheck := func(ctx context.Context, checkID uuid.UUID, schemaName string, errMsg string) {
		repo := persistence.NewCheckPostgresRepository(db, schemaName)
		check, err := repo.GetByID(ctx, checkID)
		if err != nil || check == nil {
			return
		}
		if check.Status == "success" || check.Status == "error" {
			return
		}
		check.Status = "error"
		check.ErrorMessage = errMsg
		if updateErr := repo.Update(ctx, check); updateErr != nil {
			logger.Error("failCheck: failed to update check status", zap.Error(updateErr), zap.String("check_id", checkID.String()))
			return
		}
		if onCheckDone != nil {
			onCheckDone(check.PageID, nil) // SSE push handled by Module
		}
	}

	return workers.NewWorkerPool(snapshotPort, bufferSize, failCheck)
}
