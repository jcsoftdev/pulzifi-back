package createreport

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/repositories"
)

// Handler handles the create_report use case.
type Handler struct {
	repo repositories.ReportRepository
}

// NewHandler creates a new create report handler.
func NewHandler(repo repositories.ReportRepository) *Handler {
	return &Handler{repo: repo}
}

// Handle creates a new report.
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	report := &entities.Report{
		ID:         uuid.New(),
		PageID:     req.PageID,
		Title:      req.Title,
		ReportDate: req.ReportDate,
		Content:    req.Content,
		PDFURL:     req.PDFURL,
		CreatedBy:  req.CreatedBy,
		CreatedAt:  time.Now(),
	}

	if err := h.repo.Create(ctx, report); err != nil {
		return nil, err
	}

	return &Response{Report: report}, nil
}
