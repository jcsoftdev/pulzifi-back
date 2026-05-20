package createportalsession

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	billingmocks "github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services/mocks"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/persistence/inmem"
)

func TestCreatePortalSessionHandler_Handle(t *testing.T) {
	const (
		portalURL  = "https://billing.stripe.com/p/session_test"
		customerID = "cus_test_abc"
		returnURL  = "https://app.example.com/settings/billing"
	)

	orgID := uuid.New()

	t.Run("returns portal URL when org has a stripe customer", func(t *testing.T) {
		customerRepo := inmem.NewCustomerRepo()
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
		customerRepo := inmem.NewCustomerRepo() // empty — org not registered
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
		customerRepo := inmem.NewCustomerRepo()
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
