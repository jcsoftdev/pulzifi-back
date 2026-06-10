package listplans

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
)

// planCatalogReader is the narrow read interface this handler requires.
// PlanPostgresRepository satisfies it automatically.
type planCatalogReader interface {
	ListForCatalog(ctx context.Context) ([]*entities.CatalogPlan, error)
}

// Handler lists all active, non-trial plans for the plan-catalog API.
type Handler struct {
	repo planCatalogReader
}

// NewHandler returns a Handler with the repo wired.
func NewHandler(repo planCatalogReader) *Handler {
	return &Handler{repo: repo}
}

// Handle runs the list_plans use case.
func (h *Handler) Handle(ctx context.Context) (*Response, error) {
	catalog, err := h.repo.ListForCatalog(ctx)
	if err != nil {
		return nil, err
	}

	plans := make([]PlanResponse, 0, len(catalog))
	for _, p := range catalog {
		plans = append(plans, PlanResponse{
			ID:                       p.ID,
			Code:                     p.Code,
			Name:                     p.Name,
			Description:              p.Description,
			PriceMonthlyCents:        p.PriceMonthlyCents,
			PriceYearlyCents:         p.PriceYearlyCents,
			Currency:                 p.Currency,
			ChecksAllowedMonthly:     p.ChecksAllowedMonthly,
			StoragePeriodDays:        p.StoragePeriodDays,
			AIInsightsAllowedMonthly: p.AIInsightsAllowedMonthly,
			MaxPages:                 p.MaxPages,
			MaxWorkspaces:            p.MaxWorkspaces,
			IsPopular:                p.IsPopular,
		})
	}
	return &Response{Plans: plans}, nil
}
