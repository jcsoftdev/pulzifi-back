package listorganizationswithplans

import (
	"github.com/google/uuid"
)

// OrgItem is the per-organization DTO.
type OrgItem struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Subdomain         string    `json:"subdomain"`
	SchemaName        string    `json:"schema_name"`
	PlanCode          string    `json:"plan_code"`
	PlanName          string    `json:"plan_name"`
	ChecksAllowed     int       `json:"checks_allowed_monthly"`
	StoragePeriodDays int       `json:"storage_period_days"`
}

// Response wraps the list.
type Response struct {
	Organizations []OrgItem
}
