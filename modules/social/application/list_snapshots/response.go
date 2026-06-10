package listsnapshots

import (
	"time"

	"github.com/google/uuid"
)

// SnapshotSummary is a lightweight representation of a social snapshot for list views.
// The data field is intentionally excluded — full data is fetched on demand.
type SnapshotSummary struct {
	ID             uuid.UUID `json:"id"`
	ProfileID      uuid.UUID `json:"profile_id"`
	CapturedAt     time.Time `json:"captured_at"`
	Status         string    `json:"status"`
	Error          string    `json:"error,omitempty"`
	FollowersCount int       `json:"followers_count"`
	PostsCount     int       `json:"posts_count"`
}

// Response is the output DTO for list_snapshots.
type Response struct {
	Snapshots []SnapshotSummary `json:"snapshots"`
}
