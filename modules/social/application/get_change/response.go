package getchange

import (
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
)

// ChangeDetail is the full representation of a social change.
type ChangeDetail struct {
	ID             uuid.UUID              `json:"id"`
	ProfileID      uuid.UUID              `json:"profile_id"`
	FromSnapshotID uuid.UUID              `json:"from_snapshot_id"`
	ToSnapshotID   uuid.UUID              `json:"to_snapshot_id"`
	ChangeTypes    []string               `json:"change_types"`
	Summary        entities.ChangeSummary `json:"summary"`
	CreatedAt      time.Time              `json:"created_at"`
}

// Response is the output DTO for get_change.
// It includes the full before/after ProfileData for the frontend to render
// structured diff cards (REQ-FEED-03).
type Response struct {
	Change     *ChangeDetail         `json:"change"`
	BeforeData *entities.ProfileData `json:"before_data"`
	AfterData  *entities.ProfileData `json:"after_data"`
}
