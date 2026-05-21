package giftmonth

import "github.com/google/uuid"

// GiftedPeriod describes the newly created billing period.
type GiftedPeriod struct {
	PeriodStart   string `json:"period_start"`
	PeriodEnd     string `json:"period_end"`
	ChecksAllowed int    `json:"checks_allowed"`
	NextRefillAt  string `json:"next_refill_at"`
}

// Response is the gift month response.
type Response struct {
	OrganizationID uuid.UUID    `json:"organization_id"`
	GiftedPeriod   GiftedPeriod `json:"gifted_period"`
	Message        string       `json:"message"`
}
