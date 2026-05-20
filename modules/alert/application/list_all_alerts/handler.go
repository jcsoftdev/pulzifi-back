package listallalerts

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/alert/domain/repositories"
)

type ListAllAlertsHandler struct {
	repo repositories.AlertRepository
}

func NewListAllAlertsHandler(repo repositories.AlertRepository) *ListAllAlertsHandler {
	return &ListAllAlertsHandler{repo: repo}
}

func (h *ListAllAlertsHandler) Handle(ctx context.Context, limit int) (*ListAllAlertsResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	alerts, err := h.repo.ListAll(ctx, limit)
	if err != nil {
		return nil, err
	}
	total, err := h.repo.CountAll(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]AlertItem, 0, len(alerts))
	for _, a := range alerts {
		items = append(items, AlertItem{
			ID:          a.ID.String(),
			WorkspaceID: a.WorkspaceID.String(),
			PageID:      a.PageID.String(),
			CheckID:     a.CheckID.String(),
			Title:       a.Title,
			Description: a.Description,
			Type:        a.Type,
			Read:        a.ReadAt != nil,
			CreatedAt:   a.CreatedAt.Format("2006-01-02T15:04:05Z"),
			PageName:    a.PageName,
			PageURL:     a.PageURL,
		})
	}
	return &ListAllAlertsResponse{Data: items, Total: total}, nil
}
