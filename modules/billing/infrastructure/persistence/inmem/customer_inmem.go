package inmem

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
)

// compile-time interface check
var _ repositories.CustomerRepository = (*CustomerRepo)(nil)

// CustomerRepo is a thread-safe in-memory implementation of CustomerRepository.
// It is intended for use in unit tests only.
type CustomerRepo struct {
	mu         sync.RWMutex
	byOrg      map[uuid.UUID]*entities.Customer // keyed by OrgID
	byStripeID map[string]*entities.Customer    // keyed by StripeCustomerID
}

// NewCustomerRepo returns an empty CustomerRepo.
func NewCustomerRepo() *CustomerRepo {
	return &CustomerRepo{
		byOrg:      make(map[uuid.UUID]*entities.Customer),
		byStripeID: make(map[string]*entities.Customer),
	}
}

// FindByOrgID returns the customer mapping for the given org, or nil if none.
func (r *CustomerRepo) FindByOrgID(ctx context.Context, orgID uuid.UUID) (*entities.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.byOrg[orgID]
	if !ok {
		return nil, nil
	}
	cp := *c
	return &cp, nil
}

// FindByStripeCustomerID looks up by Stripe customer identifier.
func (r *CustomerRepo) FindByStripeCustomerID(ctx context.Context, customerID string) (*entities.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.byStripeID[customerID]
	if !ok {
		return nil, nil
	}
	cp := *c
	return &cp, nil
}

// Save persists a new customer mapping.
func (r *CustomerRepo) Save(ctx context.Context, customer *entities.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *customer
	r.byOrg[customer.OrgID] = &cp
	r.byStripeID[customer.StripeCustomerID] = &cp
	return nil
}
