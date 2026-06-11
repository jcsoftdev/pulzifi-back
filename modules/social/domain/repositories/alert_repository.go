package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
)

// AlertRepository defines persistence operations for in-app SocialAlerts.
type AlertRepository interface {
	// Save persists a new SocialAlert.
	Save(ctx context.Context, alert *entities.SocialAlert) error

	// ListByWorkspace returns alerts for a workspace, newest first, paginated.
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*entities.SocialAlert, error)

	// MarkRead flips a single alert's is_read to true. No-op if not found.
	MarkRead(ctx context.Context, alertID uuid.UUID) error

	// UnreadCount returns the number of unread alerts for a workspace.
	UnreadCount(ctx context.Context, workspaceID uuid.UUID) (int, error)
}
