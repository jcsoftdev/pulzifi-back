package changepassword

import "github.com/google/uuid"

// Request carries the user ID and passwords.
type Request struct {
	UserID          uuid.UUID
	CurrentPassword string
	NewPassword     string
}
