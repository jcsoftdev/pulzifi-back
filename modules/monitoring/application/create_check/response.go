package createcheck

import (
	"time"

	"github.com/google/uuid"
)

type CreateCheckResponse struct {
	ID             uuid.UUID `json:"id"`
	PageID         uuid.UUID `json:"page_id"`
	Status         string    `json:"status"`
	ChangeDetected bool      `json:"change_detected"`
	ChangeType     string    `json:"change_type"`
	CheckedAt      time.Time `json:"checked_at"`
}
