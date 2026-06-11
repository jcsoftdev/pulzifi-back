package listalerts

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
)

// Handler lists in-app social alerts for a workspace.
type Handler struct {
	alerts repositories.AlertRepository
}

// NewHandler creates a list_alerts Handler.
func NewHandler(alerts repositories.AlertRepository) *Handler {
	return &Handler{alerts: alerts}
}

// Handle returns the workspace's alerts, newest first.
func (h *Handler) Handle(ctx context.Context, workspaceID uuid.UUID, limit, offset int) (*Response, error) {
	rows, err := h.alerts.ListByWorkspace(ctx, workspaceID, limit, offset)
	if err != nil {
		return nil, err
	}
	resp := &Response{Alerts: make([]AlertResponse, 0, len(rows))}
	for _, a := range rows {
		resp.Alerts = append(resp.Alerts, toResponse(a))
	}
	return resp, nil
}

func toResponse(a *entities.SocialAlert) AlertResponse {
	cts := make([]string, len(a.ChangeTypes))
	for i, ct := range a.ChangeTypes {
		cts[i] = string(ct)
	}
	return AlertResponse{
		ID:          a.ID,
		ProfileID:   a.ProfileID,
		WorkspaceID: a.WorkspaceID,
		ChangeID:    a.ChangeID,
		AlertType:   a.AlertType,
		Title:       a.Title,
		Description: a.Description,
		ChangeTypes: cts,
		IsRead:      a.IsRead,
		CreatedAt:   a.CreatedAt,
	}
}
