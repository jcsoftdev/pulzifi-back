package listreports

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/repositories"
)

// Handler handles the list_reports use case.
type Handler struct {
	repo repositories.ReportRepository
}

// NewHandler creates a new list reports handler.
func NewHandler(repo repositories.ReportRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle returns reports, optionally filtered by page.
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	var reports []*entities.Report
	var err error

	if req.PageID != nil {
		reports, err = h.repo.ListByPage(ctx, *req.PageID)
	} else {
		reports, err = h.repo.List(ctx)
	}
	if err != nil {
		return nil, err
	}

	if reports == nil {
		reports = []*entities.Report{}
	}

	return &Response{Reports: reports, Count: len(reports)}, nil
}
