package assignplan

import "github.com/google/uuid"

// Request carries the assignment parameters.
type Request struct {
	OrgID       uuid.UUID
	PlanCode    string
	ActorUserID string
}
