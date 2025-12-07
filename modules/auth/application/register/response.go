package register

import (
	"github.com/google/uuid"
)

// Response contains the registration response data
type Response struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Message   string    `json:"message"`
}
