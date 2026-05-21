package assignplan

import "github.com/google/uuid"

// Response is returned after successful plan assignment.
type Response struct {
	OrganizationID       uuid.UUID `json:"organization_id"`
	PlanCode             string    `json:"plan_code"`
	ChecksAllowedMonthly int       `json:"checks_allowed_monthly"`
}
