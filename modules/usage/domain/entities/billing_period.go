package entities

import "time"

// BillingPeriod represents the computed start and end of a billing cycle.
type BillingPeriod struct {
	Start time.Time
	End   time.Time
}
