package entities

import (
	"time"

	"github.com/google/uuid"
)

// Integration represents a third-party service integration (Slack, Teams, Discord, etc.)
type Integration struct {
	ID          uuid.UUID
	ServiceType string // "slack", "teams", "discord", "google_sheets"
	Config      map[string]interface{}
	Enabled     bool
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

func NewIntegration(serviceType string, config map[string]interface{}, createdBy uuid.UUID) *Integration {
	return &Integration{
		ID:          uuid.New(),
		ServiceType: serviceType,
		Config:      config,
		Enabled:     true,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
