package createcheck

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/repositories"
)

type CreateCheckHandler struct {
	repo repositories.CheckRepository
}

func NewCreateCheckHandler(repo repositories.CheckRepository) *CreateCheckHandler {
	return &CreateCheckHandler{repo: repo}
}

func (h *CreateCheckHandler) Handle(ctx context.Context, req *CreateCheckRequest) (*CreateCheckResponse, error) {
	// Create check entity
	check := entities.NewCheck(req.PageID, req.Status, req.ChangeDetected)
	check.ChangeType = req.ChangeType
	check.ScreenshotURL = req.ScreenshotURL
	check.HTMLSnapshotURL = req.HTMLSnapshotURL
	check.ErrorMessage = req.ErrorMessage
	check.DurationMs = req.DurationMs

	// Save to database
	if err := h.repo.Create(ctx, check); err != nil {
		return nil, err
	}

	// Return response
	return &CreateCheckResponse{
		ID:             check.ID,
		PageID:         check.PageID,
		Status:         check.Status,
		ChangeDetected: check.ChangeDetected,
		ChangeType:     check.ChangeType,
		CheckedAt:      check.CheckedAt,
	}, nil
}
