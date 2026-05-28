package reconcilesubscription

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	billingmocks "github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services/mocks"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/persistence/inmem"
)

func TestReconcile_AppliesActiveSub_AndProcessesDeferred(t *testing.T) {
	const (
		customerID = "cus_recon_001"
		subID      = "sub_recon_a"
		priceID    = "price_pro_monthly"
		eventID    = "evt_recon_def_1"
	)
	orgID := uuid.New()
	periodEnd := time.Now().Add(30 * 24 * time.Hour).Unix()

	gw := &billingmocks.MockStripeGateway{
		ListSubscriptionsResult: []services.StripeSubscription{
			{
				ID:               subID,
				Status:           "active",
				CurrentPeriodEnd: periodEnd,
				CustomerID:       customerID,
				PriceID:          priceID,
			},
			{
				ID:         "sub_recon_b",
				Status:     "canceled", // must be skipped
				CustomerID: customerID,
			},
		},
	}
	pa := &billingmocks.MockPlanAssigner{}
	whr := inmem.NewWebhookEventRepo()
	_, _ = whr.Save(context.Background(), &entities.WebhookEvent{
		EventID:    eventID,
		EventType:  "invoice.paid",
		ReceivedAt: time.Now().Add(-time.Hour),
		Status:     entities.WebhookEventDeferred,
		RawPayload: []byte(`{"data":{"object":{"customer":"` + customerID + `"}}}`),
	})

	h := NewHandler(gw, pa, whr)
	res, err := h.Handle(context.Background(), Input{OrgID: orgID, StripeCustomerID: customerID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.SubscriptionsFound != 2 {
		t.Errorf("subscriptions found: want 2, got %d", res.SubscriptionsFound)
	}
	if res.SubscriptionsApplied != 1 {
		t.Errorf("subscriptions applied: want 1 (only active), got %d", res.SubscriptionsApplied)
	}
	if res.DeferredEventsProcessed != 1 {
		t.Errorf("deferred events processed: want 1, got %d", res.DeferredEventsProcessed)
	}
	if pa.AssignCalls != 1 {
		t.Errorf("PlanAssigner.Assign calls: want 1, got %d", pa.AssignCalls)
	}
	if pa.LastAssignIn.StripePriceID != priceID {
		t.Errorf("priceID: want %q, got %q", priceID, pa.LastAssignIn.StripePriceID)
	}
	if pa.LastAssignIn.OrgID != orgID {
		t.Errorf("orgID: want %v, got %v", orgID, pa.LastAssignIn.OrgID)
	}
}

func TestReconcile_NoSubscriptions_StillSucceeds(t *testing.T) {
	gw := &billingmocks.MockStripeGateway{
		ListSubscriptionsResult: []services.StripeSubscription{},
	}
	pa := &billingmocks.MockPlanAssigner{}
	whr := inmem.NewWebhookEventRepo()

	h := NewHandler(gw, pa, whr)
	res, err := h.Handle(context.Background(), Input{OrgID: uuid.New(), StripeCustomerID: "cus_empty"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pa.AssignCalls != 0 {
		t.Errorf("PlanAssigner should not be called when Stripe returns no subs, got %d calls", pa.AssignCalls)
	}
	if res.SubscriptionsFound != 0 {
		t.Errorf("subscriptions found: want 0, got %d", res.SubscriptionsFound)
	}
}

func TestReconcile_MissingCustomerID_ReturnsError(t *testing.T) {
	h := NewHandler(&billingmocks.MockStripeGateway{}, &billingmocks.MockPlanAssigner{}, inmem.NewWebhookEventRepo())
	_, err := h.Handle(context.Background(), Input{OrgID: uuid.New(), StripeCustomerID: ""})
	if err == nil {
		t.Fatal("expected error when customerID is empty")
	}
}

func TestReconcile_StripeListError_Propagates(t *testing.T) {
	listErr := errors.New("stripe rate limited")
	gw := &billingmocks.MockStripeGateway{ListSubscriptionsErr: listErr}
	h := NewHandler(gw, &billingmocks.MockPlanAssigner{}, inmem.NewWebhookEventRepo())
	_, err := h.Handle(context.Background(), Input{OrgID: uuid.New(), StripeCustomerID: "cus_x"})
	if !errors.Is(err, listErr) {
		t.Fatalf("expected listErr, got: %v", err)
	}
}

func TestReconcile_AssignFailure_DoesNotHaltOtherSubs(t *testing.T) {
	const customerID = "cus_partial"
	gw := &billingmocks.MockStripeGateway{
		ListSubscriptionsResult: []services.StripeSubscription{
			{ID: "sub_a", Status: "active", CustomerID: customerID, PriceID: "price_a", CurrentPeriodEnd: time.Now().Unix()},
			{ID: "sub_b", Status: "active", CustomerID: customerID, PriceID: "price_b", CurrentPeriodEnd: time.Now().Unix()},
		},
	}
	callCount := 0
	pa := &billingmocks.MockPlanAssigner{
		AssignFn: func(_ context.Context, _ services.AssignInput) error {
			callCount++
			if callCount == 1 {
				return errors.New("first sub fails")
			}
			return nil
		},
	}
	whr := inmem.NewWebhookEventRepo()
	h := NewHandler(gw, pa, whr)
	res, err := h.Handle(context.Background(), Input{OrgID: uuid.New(), StripeCustomerID: customerID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.SubscriptionsApplied != 1 {
		t.Errorf("expected 1 applied (2nd sub succeeded), got %d", res.SubscriptionsApplied)
	}
	if callCount != 2 {
		t.Errorf("expected 2 Assign calls (1 failed, 1 succeeded), got %d", callCount)
	}
}
