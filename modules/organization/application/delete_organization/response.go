package deleteorganization

import "github.com/google/uuid"

// Response carries the result of a successful org deletion cascade.
type Response struct {
	OrgID          uuid.UUID
	Schema         string
	DeletedUserIDs []uuid.UUID
}
