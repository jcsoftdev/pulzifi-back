package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
)

// MonitoringConfigRepository defines operations for managing monitoring configs
type MonitoringConfigRepository interface {
	Create(ctx context.Context, config *entities.MonitoringConfig) error
	GetByPageID(ctx context.Context, pageID uuid.UUID) (*entities.MonitoringConfig, error)
	Update(ctx context.Context, config *entities.MonitoringConfig) error
	BulkUpdateFrequency(ctx context.Context, pageIDs []uuid.UUID, frequency string) error
	GetDueSnapshotTasks(ctx context.Context) ([]entities.SnapshotTask, error)
	GetPageURL(ctx context.Context, pageID uuid.UUID) (string, error)
	UpdateLastCheckedAt(ctx context.Context, pageID uuid.UUID) error
	MarkPageDueNow(ctx context.Context, pageID uuid.UUID) error
	GetLastCheckedAt(ctx context.Context, pageID uuid.UUID) (*time.Time, error)

	// Queue-mode methods (SCHEDULER_MODE=queue only; dormant in poll mode).

	// GetDueSnapshotTasksQueue claims pages WHERE next_run_at <= NOW() using
	// FOR UPDATE SKIP LOCKED and in the same UPDATE atomically advances
	// next_run_at = NOW() + interval so concurrent instances never double-claim.
	GetDueSnapshotTasksQueue(ctx context.Context) ([]entities.SnapshotTask, error)

	// UpdateNextRunAt sets next_run_at for a page based on its current
	// check_frequency (used after a check completes in queue mode).
	UpdateNextRunAt(ctx context.Context, pageID uuid.UUID) error
}
