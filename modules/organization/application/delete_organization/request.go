package deleteorganization

import "github.com/google/uuid"

// Request carries the input for the delete_organization use case.
type Request struct {
	OrgID     uuid.UUID // target org
	ActorID   uuid.UUID // who triggered it
	ActorType string    // "owner" | "super_admin"
}
