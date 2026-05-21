package createportalsession

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	billingmocks "github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services/mocks"
)

// fakeCustomerRepo is a local in-memory CustomerRepository for this test file.
// It avoids importing modules/billing/infrastructure/persistence/inmem from the
// application layer (which violates the application→infrastructure boundary).
type fakeCustomerRepo struct {
	mu         sync.RWMutex
	byOrg      map[uuid.UUID]*entities.Customer
	byStripeID map[string]*entities.Customer
}

func newFakeCustomerRepo() *fakeCustomerRepo {
	return &fakeCustomerRepo{
		byOrg:      make(map[uuid.UUID]*entities.Customer),
		byStripeID: make(map[string]*entities.Customer),
	}
}

func (r *fakeCustomerRepo) FindByOrgID(_ context.Context, orgID uuid.UUID) (*entities.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if c, ok := r.byOrg[orgID]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, nil
}

func (r *fakeCustomerRepo) FindByStripeCustomerID(_ context.Context, customerID string) (*entities.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if c, ok := r.byStripeID[customerID]; ok {
		cp := *c
		return &cp, nil
	}
	return nil, nil
}

func (r *fakeCustomerRepo) Save(_ context.Context, customer *entities.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *customer
	r.byOrg[customer.OrgID] = &cp
	r.byStripeID[customer.StripeCustomerID] = &cp
	return nil
}

func TestCreatePortalSessionHandler_Handle(t *testing.T) {
	const (
		portalURL  = "https://billing.stripe.com/p/session_test"
		customerID = "cus_test_abc"
		returnURL  = "https://app.example.com/settings/billing"
	)

	orgID := uuid.New()

	t.Run("returns portal URL when org has a stripe customer", func(t *testing.T) {
		customerRepo := newFakeCustomerRepo()
		_ = customerRepo.Save(context.Background(), &entities.Customer{
			OrgID:            orgID,
			StripeCustomerID: customerID,
		})

		gw := &billingmocks.MockStripeGateway{
			CreatePortalSessionResult: portalURL,
		}
		h := NewHandler(gw, customerRepo)

		resp, err := h.Handle(context.Background(), Request{
			OrgID:     orgID.String(),
			ReturnURL: returnURL,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.PortalURL != portalURL {
			t.Errorf("expected URL %q, got %q", portalURL, resp.PortalURL)
		}
		if gw.CreatePortalSessionCalls != 1 {
			t.Errorf("expected 1 CreatePortalSession call, got %d", gw.CreatePortalSessionCalls)
		}
	})

	t.Run("returns ErrNoStripeCustomer when org has no customer record", func(t *testing.T) {
		customerRepo := newFakeCustomerRepo() // empty — org not registered
		gw := &billingmocks.MockStripeGateway{}
		h := NewHandler(gw, customerRepo)

		_, err := h.Handle(context.Background(), Request{
			OrgID:     orgID.String(),
			ReturnURL: returnURL,
		})

		if !errors.Is(err, ErrNoStripeCustomer) {
			t.Errorf("expected ErrNoStripeCustomer, got %v", err)
		}
		if gw.CreatePortalSessionCalls != 0 {
			t.Errorf("expected no CreatePortalSession calls, got %d", gw.CreatePortalSessionCalls)
		}
	})

	t.Run("propagates gateway error from CreatePortalSession", func(t *testing.T) {
		customerRepo := newFakeCustomerRepo()
		_ = customerRepo.Save(context.Background(), &entities.Customer{
			OrgID:            orgID,
			StripeCustomerID: customerID,
		})

		gatewayErr := errors.New("stripe: portal error")
		gw := &billingmocks.MockStripeGateway{
			CreatePortalSessionErr: gatewayErr,
		}
		h := NewHandler(gw, customerRepo)

		_, err := h.Handle(context.Background(), Request{
			OrgID:     orgID.String(),
			ReturnURL: returnURL,
		})

		if !errors.Is(err, gatewayErr) {
			t.Errorf("expected gateway error, got %v", err)
		}
	})
}
