package listorganizationswithplans

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories"
)

// Handler handles the list_organizations_with_plans use case.
type Handler struct {
	orgPlans repositories.OrganizationPlanRepository
}

// NewHandler creates a new list organizations with plans handler.
func NewHandler(orgPlans repositories.OrganizationPlanRepository) *Handler {
	return &Handler{orgPlans: orgPlans}
}

// Handle returns all organizations with their active plan details.
func (h *Handler) Handle(ctx context.Context) (*Response, error) {
	list, err := h.orgPlans.ListOrgsWithPlans(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]OrgItem, 0, len(list))
	for _, o := range list {
		items = append(items, OrgItem{
			ID:                o.ID,
			Name:              o.Name,
			Subdomain:         o.Subdomain,
			SchemaName:        o.SchemaName,
			PlanCode:          o.PlanCode,
			PlanName:          o.PlanName,
			ChecksAllowed:     o.ChecksAllowed,
			StoragePeriodDays: o.StoragePeriodDays,
		})
	}
	return &Response{Organizations: items}, nil
}
