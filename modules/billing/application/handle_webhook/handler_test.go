package handlewebhook

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	billingmocks "github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services/mocks"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/persistence/inmem"
)

// Ensure domain sentinel is importable from the application layer.
var _ = services.ErrPlanNotFound

// ── JSON helpers ──────────────────────────────────────────────────────────────

func checkoutData(customerID, subID string) []byte {
	d, _ := json.Marshal(map[string]interface{}{
		"customer":     customerID,
		"subscription": subID,
	})
	return d
}

func subData(customerID, subID, priceID, status string, periodEnd int64) []byte {
	d, _ := json.Marshal(map[string]interface{}{
		"id":                 subID,
		"customer":           customerID,
		"status":             status,
		"current_period_end": periodEnd,
		"items": map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{
					"price": map[string]interface{}{"id": priceID},
				},
			},
		},
	})
	return d
}

func invData(customerID, subID string, periodEnd int64) []byte {
	d, _ := json.Marshal(map[string]interface{}{
		"customer":     customerID,
		"subscription": subID,
		"lines": map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{
					"period": map[string]interface{}{"end": periodEnd},
				},
			},
		},
	})
	return d
}

func makeEvent(id, typ string, data []byte) services.StripeEvent {
	return services.StripeEvent{ID: id, Type: typ, RawData: data}
}

func makeSubResult(customerID, subID, priceID, status string, periodEnd int64) services.StripeSubscription {
	return services.StripeSubscription{
		ID:               subID,
		Status:           status,
		CurrentPeriodEnd: periodEnd,
		CustomerID:       customerID,
		PriceID:          priceID,
	}
}

// ── Test ──────────────────────────────────────────────────────────────────────

func TestHandleWebhookHandler_Handle(t *testing.T) {
	const (
		secret     = "whsec_test"
		customerID = "cus_test_abc"
		subID      = "sub_test_123"
		priceID    = "price_monthly_pro"
	)

	orgID := uuid.New()
	periodEnd := time.Now().Add(30 * 24 * time.Hour).Unix()

	// Pre-seeded customer repo with org <-> customer mapping
	repoWithCustomer := func() *inmem.CustomerRepo {
		r := inmem.NewCustomerRepo()
		_ = r.Save(context.Background(), &entities.Customer{
			OrgID:            orgID,
			StripeCustomerID: customerID,
		})
		return r
	}

	// Pre-seeded webhook repo with already-processed event_id
	repoWithEvent := func(eventID string) *inmem.WebhookEventRepo {
		r := inmem.NewWebhookEventRepo()
		_, _ = r.Save(context.Background(), &entities.WebhookEvent{
			EventID:   eventID,
			EventType: "any",
			Status:    entities.WebhookEventProcessed,
		})
		return r
	}

	tests := []struct {
		name         string
		rawBody      []byte
		sig          string
		gw           *billingmocks.MockStripeGateway
		customerRepo *inmem.CustomerRepo
		webhookRepo  *inmem.WebhookEventRepo
		subRepo      *inmem.SubscriptionRepo
		assignErr    error // error MockPlanAssigner.Assign should return
		wantErr      error
		assertFn     func(t *testing.T, pa *billingmocks.MockPlanAssigner, whr *inmem.WebhookEventRepo)
	}{
		{
			name:    "checkout.session.completed — first delivery assigns plan",
			rawBody: []byte(`{}`),
			sig:     "valid-sig",
			gw: &billingmocks.MockStripeGateway{
				ConstructEventResult:       makeEvent("evt_001", "checkout.session.completed", checkoutData(customerID, subID)),
				RetrieveSubscriptionResult: makeSubResult(customerID, subID, priceID, "active", periodEnd),
			},
			customerRepo: inmem.NewCustomerRepo(), // no existing customer — will be created
			webhookRepo:  inmem.NewWebhookEventRepo(),
			subRepo:      inmem.NewSubscriptionRepo(),
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *inmem.WebhookEventRepo) {
				if pa.AssignCalls != 1 {
					t.Errorf("expected 1 Assign call, got %d", pa.AssignCalls)
				}
			},
		},
		{
			name:    "checkout.session.completed — duplicate event is no-op",
			rawBody: []byte(`{}`),
			sig:     "valid-sig",
			gw: &billingmocks.MockStripeGateway{
				ConstructEventResult: makeEvent("evt_001", "checkout.session.completed", checkoutData(customerID, subID)),
			},
			customerRepo: repoWithCustomer(),
			webhookRepo:  repoWithEvent("evt_001"), // already processed
			subRepo:      inmem.NewSubscriptionRepo(),
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *inmem.WebhookEventRepo) {
				if pa.AssignCalls != 0 {
					t.Errorf("duplicate event: expected 0 Assign calls, got %d", pa.AssignCalls)
				}
			},
		},
		{
			name:    "invoice.paid — calls PlanAssigner with refreshed period",
			rawBody: []byte(`{}`),
			sig:     "valid-sig",
			gw: &billingmocks.MockStripeGateway{
				ConstructEventResult:       makeEvent("evt_002", "invoice.paid", invData(customerID, subID, periodEnd)),
				RetrieveSubscriptionResult: makeSubResult(customerID, subID, priceID, "active", periodEnd),
			},
			customerRepo: repoWithCustomer(),
			webhookRepo:  inmem.NewWebhookEventRepo(),
			subRepo:      inmem.NewSubscriptionRepo(),
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *inmem.WebhookEventRepo) {
				if pa.AssignCalls != 1 {
					t.Errorf("invoice.paid: expected 1 Assign call, got %d", pa.AssignCalls)
				}
			},
		},
		{
			name:    "invoice.paid — duplicate is no-op",
			rawBody: []byte(`{}`),
			sig:     "valid-sig",
			gw: &billingmocks.MockStripeGateway{
				ConstructEventResult: makeEvent("evt_002", "invoice.paid", invData(customerID, subID, periodEnd)),
			},
			customerRepo: repoWithCustomer(),
			webhookRepo:  repoWithEvent("evt_002"),
			subRepo:      inmem.NewSubscriptionRepo(),
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *inmem.WebhookEventRepo) {
				if pa.AssignCalls != 0 {
					t.Errorf("dup invoice.paid: expected 0 Assign calls, got %d", pa.AssignCalls)
				}
			},
		},
		{
			name:    "customer.subscription.updated — applies new plan",
			rawBody: []byte(`{}`),
			sig:     "valid-sig",
			gw: &billingmocks.MockStripeGateway{
				ConstructEventResult: makeEvent("evt_003", "customer.subscription.updated",
					subData(customerID, subID, priceID, "active", periodEnd)),
			},
			customerRepo: repoWithCustomer(),
			webhookRepo:  inmem.NewWebhookEventRepo(),
			subRepo:      inmem.NewSubscriptionRepo(),
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *inmem.WebhookEventRepo) {
				if pa.AssignCalls != 1 {
					t.Errorf("subscription.updated: expected 1 Assign call, got %d", pa.AssignCalls)
				}
				if pa.LastAssignIn.StripePriceID != priceID {
					t.Errorf("expected priceID %q, got %q", priceID, pa.LastAssignIn.StripePriceID)
				}
			},
		},
		{
			name:    "customer.subscription.deleted — downgrades to starter (empty priceID)",
			rawBody: []byte(`{}`),
			sig:     "valid-sig",
			gw: &billingmocks.MockStripeGateway{
				ConstructEventResult: makeEvent("evt_004", "customer.subscription.deleted",
					subData(customerID, subID, priceID, "canceled", periodEnd)),
			},
			customerRepo: repoWithCustomer(),
			webhookRepo:  inmem.NewWebhookEventRepo(),
			subRepo:      inmem.NewSubscriptionRepo(),
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *inmem.WebhookEventRepo) {
				if pa.AssignCalls != 1 {
					t.Errorf("subscription.deleted: expected 1 Assign call, got %d", pa.AssignCalls)
				}
				if pa.LastAssignIn.StripePriceID != "" {
					t.Errorf("expected empty priceID for starter downgrade, got %q", pa.LastAssignIn.StripePriceID)
				}
				if pa.LastAssignIn.BillingStatus != entities.BillingCanceled {
					t.Errorf("expected BillingCanceled, got %v", pa.LastAssignIn.BillingStatus)
				}
			},
		},
		{
			name:    "invoice.payment_failed — sets PaymentStatus grace_period",
			rawBody: []byte(`{}`),
			sig:     "valid-sig",
			gw: &billingmocks.MockStripeGateway{
				ConstructEventResult:       makeEvent("evt_005", "invoice.payment_failed", invData(customerID, subID, periodEnd)),
				RetrieveSubscriptionResult: makeSubResult(customerID, subID, priceID, "past_due", periodEnd),
			},
			customerRepo: repoWithCustomer(),
			webhookRepo:  inmem.NewWebhookEventRepo(),
			subRepo:      inmem.NewSubscriptionRepo(),
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *inmem.WebhookEventRepo) {
				if pa.AssignCalls != 1 {
					t.Errorf("payment_failed: expected 1 Assign call, got %d", pa.AssignCalls)
				}
				if pa.LastAssignIn.PaymentStatus != "grace_period" {
					t.Errorf("expected PaymentStatus grace_period, got %q", pa.LastAssignIn.PaymentStatus)
				}
			},
		},
		{
			name:    "customer.subscription.updated — unknown price_id is silent no-op (200)",
			rawBody: []byte(`{}`),
			sig:     "valid-sig",
			gw: &billingmocks.MockStripeGateway{
				ConstructEventResult: makeEvent("evt_006", "customer.subscription.updated",
					subData(customerID, subID, "price_unknown_xyz", "active", periodEnd)),
			},
			customerRepo: repoWithCustomer(),
			webhookRepo:  inmem.NewWebhookEventRepo(),
			subRepo:      inmem.NewSubscriptionRepo(),
			// PlanAssigner returns ErrPlanNotFound for unknown price_id.
			// Handler must treat this as no-op (nil return) per FR5.
			assignErr: services.ErrPlanNotFound,
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *inmem.WebhookEventRepo) {
				if pa.AssignCalls != 1 {
					t.Errorf("unknown price_id: expected 1 Assign attempt, got %d", pa.AssignCalls)
				}
			},
		},
		{
			name:    "invalid signature — returns ErrInvalidSignature",
			rawBody: []byte(`{}`),
			sig:     "bad-sig",
			gw: &billingmocks.MockStripeGateway{
				ConstructEventErr: errors.New("stripe: invalid signature"),
			},
			customerRepo: inmem.NewCustomerRepo(),
			webhookRepo:  inmem.NewWebhookEventRepo(),
			subRepo:      inmem.NewSubscriptionRepo(),
			wantErr:      ErrInvalidSignature,
		},
		{
			name:    "missing signature header — returns ErrInvalidSignature",
			rawBody: []byte(`{}`),
			sig:     "", // empty = missing
			gw:      &billingmocks.MockStripeGateway{},
			customerRepo: inmem.NewCustomerRepo(),
			webhookRepo:  inmem.NewWebhookEventRepo(),
			subRepo:      inmem.NewSubscriptionRepo(),
			wantErr:      ErrInvalidSignature,
		},
		// MUST: unknown event types MUST be silently no-op'd without error (per spec).
		// Stripe sends many event types; the handler must not return errors for ones
		// it doesn't recognise — that would trigger retries from Stripe.
		{
			name:    "unknown event type — silent no-op, returns nil",
			rawBody: []byte(`{}`),
			sig:     "valid-sig",
			gw: &billingmocks.MockStripeGateway{
				ConstructEventResult: makeEvent("evt_tax", "customer.tax_id.created", []byte(`{}`)),
			},
			customerRepo: repoWithCustomer(),
			webhookRepo:  inmem.NewWebhookEventRepo(),
			subRepo:      inmem.NewSubscriptionRepo(),
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *inmem.WebhookEventRepo) {
				// Unknown event type must never trigger PlanAssigner.
				if pa.AssignCalls != 0 {
					t.Errorf("unknown event type: expected 0 Assign calls, got %d", pa.AssignCalls)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pa := &billingmocks.MockPlanAssigner{AssignErr: tt.assignErr}
			h := NewHandler(tt.gw, secret, pa, tt.customerRepo, tt.webhookRepo, tt.subRepo)

			err := h.Handle(context.Background(), tt.rawBody, tt.sig)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.assertFn != nil {
				tt.assertFn(t, pa, tt.webhookRepo)
			}
		})
	}
}
