package services

import (
	"context"

	"github.com/google/uuid"
)

// Scheduler defines the port for scheduling and triggering monitoring checks.
// The concrete implementation lives in infrastructure/scheduler.
type Scheduler interface {
	// WakeUp signals the scheduler to re-evaluate due tasks immediately.
	WakeUp()
	// TriggerPageCheck schedules an immediate check for a specific page.
	TriggerPageCheck(ctx context.Context, schema string, pageID uuid.UUID) error
	// Start begins the scheduler loop. Called once at startup.
	Start(ctx context.Context)
}
