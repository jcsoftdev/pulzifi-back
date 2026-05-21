package getreport

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/repositories"
)

// Handler handles the get_report use case.
type Handler struct {
	repo repositories.ReportRepository
}

// NewHandler creates a new get report handler.
func NewHandler(repo repositories.ReportRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle fetches a single report by ID.
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	report, err := h.repo.GetByID(ctx, req.ReportID)
	if err != nil {
		return nil, err
	}
	return &Response{Report: report}, nil
}
