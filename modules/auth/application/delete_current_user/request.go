package deletecurrentuser

import "github.com/google/uuid"

// Request carries the user ID to delete.
type Request struct {
	UserID uuid.UUID
}
