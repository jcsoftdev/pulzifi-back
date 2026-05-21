package entities

import (
	"time"

	"github.com/google/uuid"
)

// UsageTracking represents a billing period row in the tenant schema.
type UsageTracking struct {
	ID                uuid.UUID
	PeriodStart       time.Time
	PeriodEnd         time.Time
	ChecksAllowed     int
	ChecksUsed        int
	NextRefillAt      *time.Time // nullable
	StoragePeriodDays int
}
