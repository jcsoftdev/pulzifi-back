package login

import (
"github.com/google/uuid"
)

// Response contains the login response data
type Response struct {
	UserID       uuid.UUID `json:"user_id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
}
