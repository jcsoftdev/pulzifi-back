package markalertread

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
)

// Handler marks a single social alert as read.
type Handler struct {
	alerts repositories.AlertRepository
}

// NewHandler creates a mark_alert_read Handler.
func NewHandler(alerts repositories.AlertRepository) *Handler {
	return &Handler{alerts: alerts}
}

// Handle flips the alert's is_read to true.
func (h *Handler) Handle(ctx context.Context, alertID uuid.UUID) error {
	return h.alerts.MarkRead(ctx, alertID)
}
