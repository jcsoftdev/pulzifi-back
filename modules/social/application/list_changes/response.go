package listchanges

import (
	"time"

	"github.com/google/uuid"
)

// ChangeSummaryItem is a lightweight representation of a social change for list views.
type ChangeSummaryItem struct {
	ID             uuid.UUID `json:"id"`
	ProfileID      uuid.UUID `json:"profile_id"`
	FromSnapshotID uuid.UUID `json:"from_snapshot_id"`
	ToSnapshotID   uuid.UUID `json:"to_snapshot_id"`
	ChangeTypes    []string  `json:"change_types"`
	CreatedAt      time.Time `json:"created_at"`
}

// Response is the output DTO for list_changes.
type Response struct {
	Changes []*ChangeSummaryItem `json:"changes"`
}
