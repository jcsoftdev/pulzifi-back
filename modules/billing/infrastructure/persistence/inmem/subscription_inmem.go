package inmem

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
)

// compile-time interface check
var _ repositories.SubscriptionRepository = (*SubscriptionRepo)(nil)

// SubscriptionRepo is a thread-safe in-memory implementation of SubscriptionRepository.
// It is intended for use in unit tests only.
type SubscriptionRepo struct {
	mu   sync.RWMutex
	byID map[uuid.UUID]*entities.Subscription      // keyed by OrgID
	bySub map[string]*entities.Subscription        // keyed by StripeSubscriptionID
}

// NewSubscriptionRepo returns an empty SubscriptionRepo.
func NewSubscriptionRepo() *SubscriptionRepo {
	return &SubscriptionRepo{
		byID:  make(map[uuid.UUID]*entities.Subscription),
		bySub: make(map[string]*entities.Subscription),
	}
}

// FindByOrgID returns the subscription for the given organisation, or nil if none.
func (r *SubscriptionRepo) FindByOrgID(ctx context.Context, orgID uuid.UUID) (*entities.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sub, ok := r.byID[orgID]
	if !ok {
		return nil, nil
	}
	cp := *sub
	return &cp, nil
}

// FindByStripeSubscriptionID looks up a subscription by Stripe subscription ID.
func (r *SubscriptionRepo) FindByStripeSubscriptionID(ctx context.Context, subID string) (*entities.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sub, ok := r.bySub[subID]
	if !ok {
		return nil, nil
	}
	cp := *sub
	return &cp, nil
}

// Save inserts a new subscription record.
func (r *SubscriptionRepo) Save(ctx context.Context, sub *entities.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *sub
	r.byID[sub.OrgID] = &cp
	if sub.StripeSubscriptionID != "" {
		r.bySub[sub.StripeSubscriptionID] = &cp
	}
	return nil
}

// Update overwrites the subscription record for the given org.
func (r *SubscriptionRepo) Update(ctx context.Context, sub *entities.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Remove stale stripe sub ID index if it changed
	if old, ok := r.byID[sub.OrgID]; ok && old.StripeSubscriptionID != sub.StripeSubscriptionID {
		delete(r.bySub, old.StripeSubscriptionID)
	}
	cp := *sub
	r.byID[sub.OrgID] = &cp
	if sub.StripeSubscriptionID != "" {
		r.bySub[sub.StripeSubscriptionID] = &cp
	}
	return nil
}
