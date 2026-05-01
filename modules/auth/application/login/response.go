package login

import "github.com/google/uuid"

type Response struct {
	UserID       uuid.UUID `json:"-"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int64     `json:"expires_in"`
	Tenant       *string   `json:"tenant,omitempty"`
}
