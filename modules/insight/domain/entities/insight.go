package entities

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Insight struct {
	ID          uuid.UUID       `json:"id"`
	PageID      uuid.UUID       `json:"page_id"`
	CheckID     uuid.UUID       `json:"check_id"`
	InsightType string          `json:"insight_type"`
	Title       string          `json:"title"`
	Content     string          `json:"content"`
	Metadata    json.RawMessage `json:"metadata"`
	CreatedAt   time.Time       `json:"created_at"`
	DeletedAt   *time.Time      `json:"deleted_at,omitempty"`
}
