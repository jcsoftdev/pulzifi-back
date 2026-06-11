package services

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// InsightSummary is a flattened AI insight used as input for report
// summarization. It is intentionally decoupled from the insight module's
// entity so the report module never imports it directly.
type InsightSummary struct {
	Type      string
	Title     string
	Content   string
	CreatedAt time.Time
}

// InsightReader fetches the AI insights of a page. Implemented by a
// cross-module adapter in cmd/wiring/report.
type InsightReader interface {
	ListByPage(ctx context.Context, pageID uuid.UUID) ([]InsightSummary, error)
}

// ReportSummarizer turns a page's AI insights into a plain-text report:
// a short overview followed by bullet-point recommended actions when warranted.
type ReportSummarizer interface {
	Summarize(ctx context.Context, pageTitle string, insights []InsightSummary) (string, error)
}
