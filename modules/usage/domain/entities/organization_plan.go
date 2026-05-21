package entities

import (
	"time"

	"github.com/google/uuid"
)

// OrganizationPlan represents an organization's plan assignment in the public schema.
type OrganizationPlan struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	PlanID         uuid.UUID
	Status         string
	StartedAt      time.Time
	EndedAt        *time.Time
	CreatedBy      *uuid.UUID
}
