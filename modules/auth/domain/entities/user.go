package entities

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID                        uuid.UUID
	Email                     string
	PasswordHash              string
	FirstName                 string
	LastName                  string
	AvatarURL                 *string
	EmailVerified             bool
	EmailNotificationsEnabled bool
	NotificationFrequency     string // 'immediate', 'daily_digest', 'weekly_digest'
	CreatedAt                 time.Time
	UpdatedAt                 time.Time
	DeletedAt                 *time.Time
}

// NewUser creates a new user entity
func NewUser(email, passwordHash, firstName, lastName string) *User {
	return &User{
		ID:                        uuid.New(),
		Email:                     email,
		PasswordHash:              passwordHash,
		FirstName:                 firstName,
		LastName:                  lastName,
		EmailVerified:             false,
		EmailNotificationsEnabled: true,
		NotificationFrequency:     "immediate",
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}
}
