package createcheckoutsession

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
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

// fakePlanRepo returns canned price IDs keyed by (code, cycle).
type fakePlanRepo struct {
	monthly map[string]string
	yearly  map[string]string
	err     error
}

func (r *fakePlanRepo) FindStripePriceID(_ context.Context, code, cycle string) (string, error) {
	if r.err != nil {
		return "", r.err
	}
	switch cycle {
	case "monthly":
		if v, ok := r.monthly[code]; ok {
			return v, nil
		}
	case "yearly":
		if v, ok := r.yearly[code]; ok {
			return v, nil
		}
	}
	return "", repositories.ErrPlanNotFound
}

func TestCreateCheckoutSessionHandler_Handle(t *testing.T) {
	const (
		priceMonthly = "price_monthly_123"
		priceYearly  = "price_yearly_123"
		customerID   = "cus_test_123"
		checkoutURL  = "https://checkout.stripe.com/pay/cs_test"
	)

	newHandler := func(gateway *billingmocks.MockStripeGateway, plan repositories.PlanRepository) *Handler {
		return NewHandler(gateway, newFakeCustomerRepo(), plan)
	}

	plansPro := &fakePlanRepo{
		monthly: map[string]string{"pro": priceMonthly},
		yearly:  map[string]string{"pro": priceYearly},
	}

	t.Run("returns checkout URL for monthly billing cycle", func(t *testing.T) {
		gw := &billingmocks.MockStripeGateway{
			EnsureCustomerResult:        customerID,
			CreateCheckoutSessionResult: checkoutURL,
		}
		h := newHandler(gw, plansPro)

		resp, err := h.Handle(context.Background(), Request{
			OrgID:        "org-uuid-001",
			OrgEmail:     "admin@acme.com",
			OrgName:      "Acme",
			PlanID:       "pro",
			BillingCycle: "monthly",
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
		h := newHandler(gw, plansPro)

		resp, err := h.Handle(context.Background(), Request{
			OrgID:        "org-uuid-001",
			OrgEmail:     "admin@acme.com",
			OrgName:      "Acme",
			PlanID:       "pro",
			BillingCycle: "yearly",
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
		h := newHandler(gw, plansPro)

		_, err := h.Handle(context.Background(), Request{
			OrgID:        "org-uuid-001",
			OrgEmail:     "admin@acme.com",
			OrgName:      "Acme",
			PlanID:       "pro",
			BillingCycle: "quarterly",
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
		plans := &fakePlanRepo{
			monthly: map[string]string{"starter": ""},
			yearly:  map[string]string{"starter": ""},
		}
		h := newHandler(gw, plans)

		_, err := h.Handle(context.Background(), Request{
			OrgID:        "org-uuid-001",
			OrgEmail:     "admin@acme.com",
			OrgName:      "Acme",
			PlanID:       "starter",
			BillingCycle: "monthly",
		})

		if !errors.Is(err, ErrMissingPriceID) {
			t.Errorf("expected ErrMissingPriceID, got %v", err)
		}
	})

	t.Run("returns ErrPlanNotFound when plan code does not exist", func(t *testing.T) {
		gw := &billingmocks.MockStripeGateway{}
		h := newHandler(gw, plansPro)

		_, err := h.Handle(context.Background(), Request{
			OrgID:        "org-uuid-001",
			OrgEmail:     "admin@acme.com",
			OrgName:      "Acme",
			PlanID:       "ghost",
			BillingCycle: "monthly",
		})

		if !errors.Is(err, ErrPlanNotFound) {
			t.Errorf("expected ErrPlanNotFound, got %v", err)
		}
	})

	t.Run("propagates gateway error from EnsureCustomer", func(t *testing.T) {
		gatewayErr := errors.New("stripe: network error")
		gw := &billingmocks.MockStripeGateway{
			EnsureCustomerErr: gatewayErr,
		}
		h := newHandler(gw, plansPro)

		_, err := h.Handle(context.Background(), Request{
			OrgID:        "org-uuid-001",
			OrgEmail:     "admin@acme.com",
			OrgName:      "Acme",
			PlanID:       "pro",
			BillingCycle: "monthly",
		})

		if !errors.Is(err, gatewayErr) {
			t.Errorf("expected gateway error, got %v", err)
		}
	})
}
