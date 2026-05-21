package listplans

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories"
)

// Handler handles the list_plans use case.
type Handler struct {
	plans repositories.PlanRepository
}

// NewHandler creates a new list plans handler.
func NewHandler(plans repositories.PlanRepository) *Handler {
	return &Handler{plans: plans}
}

// Handle returns all active plans.
func (h *Handler) Handle(ctx context.Context) (*Response, error) {
	list, err := h.plans.ListActive(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]PlanItem, 0, len(list))
	for _, p := range list {
		items = append(items, fromEntity(p))
	}
	return &Response{Plans: items}, nil
}
