package updatecurrentuser

import "github.com/google/uuid"

// Request carries the user ID and updated fields.
type Request struct {
	UserID    uuid.UUID
	FirstName string
	LastName  string
	Roles     []string
}
