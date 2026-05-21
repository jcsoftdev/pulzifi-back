package createcheckoutsession

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

func TestCreateCheckoutSessionHandler_Handle(t *testing.T) {
	const (
		priceMonthly = "price_monthly_123"
		priceYearly  = "price_yearly_123"
		customerID   = "cus_test_123"
		checkoutURL  = "https://checkout.stripe.com/pay/cs_test"
	)

	newHandler := func(gateway *billingmocks.MockStripeGateway) *Handler {
		customerRepo := newFakeCustomerRepo()
		return NewHandler(gateway, customerRepo)
	}

	t.Run("returns checkout URL for monthly billing cycle", func(t *testing.T) {
		gw := &billingmocks.MockStripeGateway{
			EnsureCustomerResult:        customerID,
			CreateCheckoutSessionResult: checkoutURL,
		}
		h := newHandler(gw)

		resp, err := h.Handle(context.Background(), Request{
			OrgID:        "org-uuid-001",
			OrgEmail:     "admin@acme.com",
			OrgName:      "Acme",
			PlanID:       "plan-pro",
			BillingCycle: "monthly",
			StripePriceIDMonthly: priceMonthly,
			StripePriceIDYearly:  priceYearly,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.CheckoutURL != checkoutURL {
			t.Errorf("expected URL %q, got %q", checkoutURL, resp.CheckoutURL)
		}
		if gw.EnsureCustomerCalls != 1 {
			t.Errorf("expected 1 EnsureCustomer call, got %d", gw.EnsureCustomerCalls)
		}
		if gw.CreateCheckoutSessionCalls != 1 {
			t.Errorf("expected 1 CreateCheckoutSession call, got %d", gw.CreateCheckoutSessionCalls)
		}
	})

	t.Run("returns checkout URL for yearly billing cycle", func(t *testing.T) {
		gw := &billingmocks.MockStripeGateway{
			EnsureCustomerResult:        customerID,
			CreateCheckoutSessionResult: checkoutURL,
		}
		h := newHandler(gw)

		resp, err := h.Handle(context.Background(), Request{
			OrgID:        "org-uuid-001",
			OrgEmail:     "admin@acme.com",
			OrgName:      "Acme",
			PlanID:       "plan-pro",
			BillingCycle: "yearly",
			StripePriceIDMonthly: priceMonthly,
			StripePriceIDYearly:  priceYearly,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.CheckoutURL != checkoutURL {
			t.Errorf("expected URL %q, got %q", checkoutURL, resp.CheckoutURL)
		}
	})

	t.Run("returns ErrInvalidBillingCycle for unknown billing cycle", func(t *testing.T) {
		gw := &billingmocks.MockStripeGateway{}
		h := newHandler(gw)

		_, err := h.Handle(context.Background(), Request{
			OrgID:        "org-uuid-001",
			OrgEmail:     "admin@acme.com",
			OrgName:      "Acme",
			PlanID:       "plan-pro",
			BillingCycle: "quarterly",
			StripePriceIDMonthly: priceMonthly,
			StripePriceIDYearly:  priceYearly,
		})

		if !errors.Is(err, ErrInvalidBillingCycle) {
			t.Errorf("expected ErrInvalidBillingCycle, got %v", err)
		}
		if gw.EnsureCustomerCalls != 0 {
			t.Errorf("expected no EnsureCustomer calls, got %d", gw.EnsureCustomerCalls)
		}
	})

	t.Run("returns ErrMissingPriceID when plan has no stripe price configured", func(t *testing.T) {
		gw := &billingmocks.MockStripeGateway{}
		h := newHandler(gw)

		_, err := h.Handle(context.Background(), Request{
			OrgID:        "org-uuid-001",
			OrgEmail:     "admin@acme.com",
			OrgName:      "Acme",
			PlanID:       "plan-starter",
			BillingCycle: "monthly",
			StripePriceIDMonthly: "", // no price configured
			StripePriceIDYearly:  "",
		})

		if !errors.Is(err, ErrMissingPriceID) {
			t.Errorf("expected ErrMissingPriceID, got %v", err)
		}
	})

	t.Run("propagates gateway error from EnsureCustomer", func(t *testing.T) {
		gatewayErr := errors.New("stripe: network error")
		gw := &billingmocks.MockStripeGateway{
			EnsureCustomerErr: gatewayErr,
		}
		h := newHandler(gw)

		_, err := h.Handle(context.Background(), Request{
			OrgID:        "org-uuid-001",
			OrgEmail:     "admin@acme.com",
			OrgName:      "Acme",
			PlanID:       "plan-pro",
			BillingCycle: "monthly",
			StripePriceIDMonthly: priceMonthly,
			StripePriceIDYearly:  priceYearly,
		})

		if !errors.Is(err, gatewayErr) {
			t.Errorf("expected gateway error, got %v", err)
		}
	})
}
