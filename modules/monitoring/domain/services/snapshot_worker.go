package services

import (
	"context"

	"github.com/google/uuid"
)

// SnapshotExecutor defines the port for executing page snapshot checks.
// The concrete implementation lives in modules/snapshot/application.
// It extends workers.SnapshotPort with the callback needed for SSE push.
type SnapshotExecutor interface {
	// ExecuteCheck runs a snapshot check for the given check ID and page URL.
	ExecuteCheck(ctx context.Context, checkID uuid.UUID, targetURL string, schemaName string) error
	// SetOnCheckDone registers a callback invoked after every check completes.
	SetOnCheckDone(fn func(pageID uuid.UUID, checkJSON []byte))
}
