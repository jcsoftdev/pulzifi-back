package createreport

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/services"
)

const emptyInsightsMessage = "No AI insights have been generated for this page yet. " +
	"Once monitoring detects changes and produces insights, this report will summarize them."

// Handler handles the create_report use case.
type Handler struct {
	repo       repositories.ReportRepository
	insights   services.InsightReader    // optional — nil disables AI summarization
	summarizer services.ReportSummarizer // optional — nil disables AI summarization
}

// NewHandler creates a create report handler that stores the report as-is,
// without AI summarization.
func NewHandler(repo repositories.ReportRepository) *Handler {
	return &Handler{repo: repo}
}

// NewHandlerWithAI creates a create report handler that, on create, summarizes
// the page's AI insights into the report content (a monthly report).
func NewHandlerWithAI(repo repositories.ReportRepository, insights services.InsightReader, summarizer services.ReportSummarizer) *Handler {
	return &Handler{repo: repo, insights: insights, summarizer: summarizer}
}

// Handle creates a new report. When AI summarization is wired, it fetches the
// page's insights and writes a plain-text summary into content["summary"].
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	content := req.Content
	if content == nil {
		content = entities.Content{}
	}

	if h.insights != nil && h.summarizer != nil {
		summary, count, err := h.buildSummary(ctx, req)
		if err != nil {
			return nil, err
		}
		content["summary"] = summary
		content["insights_count"] = count
	}

	report := &entities.Report{
		ID:         uuid.New(),
		PageID:     req.PageID,
		Title:      req.Title,
		ReportDate: req.ReportDate,
		Content:    content,
		PDFURL:     req.PDFURL,
		CreatedBy:  req.CreatedBy,
		CreatedAt:  time.Now(),
	}

	if err := h.repo.Create(ctx, report); err != nil {
		return nil, err
	}

	return &Response{Report: report}, nil
}

// buildSummary fetches the page's insights and summarizes them. When the page
// has no insights, it returns a friendly placeholder instead of calling the LLM.
func (h *Handler) buildSummary(ctx context.Context, req *Request) (string, int, error) {
	insights, err := h.insights.ListByPage(ctx, req.PageID)
	if err != nil {
		return "", 0, err
	}
	if len(insights) == 0 {
		return emptyInsightsMessage, 0, nil
	}
	summary, err := h.summarizer.Summarize(ctx, req.Title, insights)
	if err != nil {
		return "", 0, err
	}
	return summary, len(insights), nil
}
