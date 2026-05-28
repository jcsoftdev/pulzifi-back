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

	// AI insight cap mirrors the checks columns. Allowed = 2147483647
	// (INT max) is the sentinel for "unlimited" — written by the PlanAssigner
	// when the active plan's ai_insights_allowed_monthly is NULL (Enterprise).
	AIInsightsAllowed int
	AIInsightsUsed    int
}
