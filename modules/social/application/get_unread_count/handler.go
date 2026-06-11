package getunreadcount

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
)

// Handler returns the unread social-alert count for a workspace.
type Handler struct {
	alerts repositories.AlertRepository
}

// NewHandler creates a get_unread_count Handler.
func NewHandler(alerts repositories.AlertRepository) *Handler {
	return &Handler{alerts: alerts}
}

// Handle returns the unread count.
func (h *Handler) Handle(ctx context.Context, workspaceID uuid.UUID) (*Response, error) {
	count, err := h.alerts.UnreadCount(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return &Response{UnreadCount: count}, nil
}
