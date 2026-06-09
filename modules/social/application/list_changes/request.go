package listchanges

import "github.com/google/uuid"

// Request is the input DTO for listing social changes.
// At least one of ProfileID or WorkspaceID should be set.
// When ProfileID is set, WorkspaceID is ignored.
type Request struct {
	ProfileID   *uuid.UUID `json:"profile_id"`
	WorkspaceID *uuid.UUID `json:"workspace_id"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
}
