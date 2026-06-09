package runcheck

import (
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
)

// Response is the output DTO for run_check.
type Response struct {
	// SnapshotID is the UUID of the persisted snapshot.
	SnapshotID uuid.UUID `json:"snapshot_id"`

	// Status is the snapshot status ("success" or "failed").
	Status string `json:"status"`

	// ChangeCreated is true when a SocialChange was persisted (diff was non-empty).
	ChangeCreated bool `json:"change_created"`

	// ChangeID is the UUID of the persisted change, when ChangeCreated is true.
	ChangeID *uuid.UUID `json:"change_id,omitempty"`

	// NextCheckAt is the updated next_check_at on the profile.
	NextCheckAt *time.Time `json:"next_check_at"`

	// Data is the captured ProfileData (nil on failure).
	Data *entities.ProfileData `json:"data,omitempty"`
}
