package listplans

import "github.com/jcsoftdev/pulzifi-back/modules/usage/domain/entities"

// PlanItem is the per-plan DTO returned to the HTTP layer.
type PlanItem struct {
	ID                   interface{} `json:"id"`
	Code                 string      `json:"code"`
	Name                 string      `json:"name"`
	Description          string      `json:"description"`
	ChecksAllowedMonthly int         `json:"checks_allowed_monthly"`
	IsActive             bool        `json:"is_active"`
	StoragePeriodDays    int         `json:"storage_period_days"`
}

// Response wraps the list of plans.
type Response struct {
	Plans []PlanItem
}

func fromEntity(p *entities.Plan) PlanItem {
	desc := ""
	if p.Description != nil {
		desc = *p.Description
	}
	return PlanItem{
		ID:                   p.ID,
		Code:                 p.Code,
		Name:                 p.Name,
		Description:          desc,
		ChecksAllowedMonthly: p.ChecksAllowedMonthly,
		IsActive:             p.IsActive,
		StoragePeriodDays:    p.StoragePeriodDays,
	}
}
