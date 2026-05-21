package entities

import "github.com/google/uuid"

// Plan represents a subscription plan in the public schema.
type Plan struct {
	ID                   uuid.UUID
	Code                 string
	Name                 string
	Description          *string // nullable
	ChecksAllowedMonthly int
	IsActive             bool
	StoragePeriodDays    int
}
