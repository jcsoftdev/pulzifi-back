package entities

import (
	"time"

	"github.com/google/uuid"
)

type IntegrationStatus string

const (
	IntegrationActive       IntegrationStatus = "active"
	IntegrationDisconnected IntegrationStatus = "disconnected"
	IntegrationExpired      IntegrationStatus = "expired"
)

type Integration struct {
	ID             uuid.UUID
	OrgID          uuid.UUID
	ServiceType    string
	Status         IntegrationStatus
	AccessToken    string                 // plaintext on entity; encrypted in DB layer
	RefreshToken   string                 // plaintext on entity, "" if none
	TokenExpiresAt *time.Time
	ProviderMeta   map[string]any
	CreatedBy      uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}
