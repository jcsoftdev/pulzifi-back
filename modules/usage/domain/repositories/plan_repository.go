package repositories

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/entities"
)

// PlanRepository manages plan records in the public schema.
type PlanRepository interface {
	// ListActive returns all active plans ordered by checks_allowed_monthly ASC.
	ListActive(ctx context.Context) ([]*entities.Plan, error)

	// GetByCode returns a single active plan by its code.
	GetByCode(ctx context.Context, code string) (*entities.Plan, error)
}
