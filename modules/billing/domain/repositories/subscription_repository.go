package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
)

// SubscriptionRepository persists Stripe subscription state linked to an organisation.
type SubscriptionRepository interface {
	// FindByOrgID returns the subscription for the given organisation, or nil if none.
	FindByOrgID(ctx context.Context, orgID uuid.UUID) (*entities.Subscription, error)

	// FindByStripeSubscriptionID looks up by the Stripe subscription identifier.
	FindByStripeSubscriptionID(ctx context.Context, subID string) (*entities.Subscription, error)

	// Save inserts a new subscription record.
	Save(ctx context.Context, sub *entities.Subscription) error

	// Update overwrites the subscription record for the given org.
	Update(ctx context.Context, sub *entities.Subscription) error
}
