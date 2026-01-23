package listinsights

import (
	"time"

	"github.com/google/uuid"
)

type InsightResponse struct {
	ID          uuid.UUID   `json:"id"`
	PageID      uuid.UUID   `json:"page_id"`
	CheckID     uuid.UUID   `json:"check_id"`
	InsightType string      `json:"insight_type"`
	Title       string      `json:"title"`
	Content     string      `json:"content"`
	Metadata    interface{} `json:"metadata"`
	CreatedAt   time.Time   `json:"created_at"`
}

type ListInsightsResponse struct {
	Insights []*InsightResponse `json:"insights"`
}
