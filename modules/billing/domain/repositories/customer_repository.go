package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
)

// CustomerRepository persists the mapping between organisations and Stripe customers.
type CustomerRepository interface {
	// FindByOrgID returns the customer mapping for the given org, or nil if none.
	FindByOrgID(ctx context.Context, orgID uuid.UUID) (*entities.Customer, error)

	// FindByStripeCustomerID looks up by Stripe customer identifier.
	FindByStripeCustomerID(ctx context.Context, customerID string) (*entities.Customer, error)

	// Save persists a new customer mapping.
	Save(ctx context.Context, customer *entities.Customer) error
}
