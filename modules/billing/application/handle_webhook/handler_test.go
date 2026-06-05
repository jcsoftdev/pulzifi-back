package handlewebhook

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	billingmocks "github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services/mocks"
)

// ── local fakes ───────────────────────────────────────────────────────────────
// These avoid importing modules/billing/infrastructure/persistence/inmem from
// the application layer (application→infrastructure boundary violation).

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

type fakeWebhookEventRepo struct {
	mu     sync.Mutex
	events map[string]*entities.WebhookEvent
}

func newFakeWebhookEventRepo() *fakeWebhookEventRepo {
	return &fakeWebhookEventRepo{events: make(map[string]*entities.WebhookEvent)}
}

func (r *fakeWebhookEventRepo) Save(_ context.Context, event *entities.WebhookEvent) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.events[event.EventID]; exists {
		return false, nil
	}
	cp := *event
	r.events[event.EventID] = &cp
	return true, nil
}

func (r *fakeWebhookEventRepo) MarkProcessed(_ context.Context, eventID string, status entities.WebhookEventStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ev, ok := r.events[eventID]; ok {
		now := time.Now()
		ev.ProcessedAt = &now
		ev.Status = status
	}
	return nil
}

func (r *fakeWebhookEventRepo) FindDeferredByCustomer(_ context.Context, _ string) ([]*entities.WebhookEvent, error) {
	return nil, nil
}

type fakeSubscriptionRepo struct {
	mu    sync.RWMutex
	byID  map[uuid.UUID]*entities.Subscription
	bySub map[string]*entities.Subscription
}

func newFakeSubscriptionRepo() *fakeSubscriptionRepo {
	return &fakeSubscriptionRepo{
		byID:  make(map[uuid.UUID]*entities.Subscription),
		bySub: make(map[string]*entities.Subscription),
	}
}

func (r *fakeSubscriptionRepo) FindByOrgID(_ context.Context, orgID uuid.UUID) (*entities.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if sub, ok := r.byID[orgID]; ok {
		cp := *sub
		return &cp, nil
	}
	return nil, nil
}

func (r *fakeSubscriptionRepo) FindByStripeSubscriptionID(_ context.Context, subID string) (*entities.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if sub, ok := r.bySub[subID]; ok {
		cp := *sub
		return &cp, nil
	}
	return nil, nil
}

func (r *fakeSubscriptionRepo) Save(_ context.Context, sub *entities.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *sub
	r.byID[sub.OrgID] = &cp
	if sub.StripeSubscriptionID != "" {
		r.bySub[sub.StripeSubscriptionID] = &cp
	}
	return nil
}

func (r *fakeSubscriptionRepo) Update(_ context.Context, sub *entities.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
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

// fakePlanRepo records UpsertStripePricing calls for pricing-sync assertions.
type fakePlanRepo struct {
	mu        sync.Mutex
	calls     int
	lastCode  string
	lastCycle string
	lastPrice string
	lastCents int64
	lastCur   string
	knownCode string // codes != knownCode return ErrPlanNotFound (when set)
}

func (r *fakePlanRepo) FindStripePriceID(_ context.Context, _, _ string) (string, error) {
	return "", nil
}

func (r *fakePlanRepo) UpsertStripePricing(_ context.Context, planCode, cycle, priceID string, amountCents int64, currency string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls++
	r.lastCode, r.lastCycle, r.lastPrice = planCode, cycle, priceID
	r.lastCents, r.lastCur = amountCents, currency
	if r.knownCode != "" && planCode != r.knownCode {
		return repositories.ErrPlanNotFound
	}
	return nil
}

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

// priceData builds a Stripe price.* event object. interval is "month" | "year".
// planCode is written to metadata.plan_code (empty → omitted).
func priceData(priceID, productID string, unitAmount int64, currency, interval, planCode string) []byte {
	obj := map[string]interface{}{
		"id":          priceID,
		"product":     productID,
		"unit_amount": unitAmount,
		"currency":    currency,
		"active":      true,
		"recurring":   map[string]interface{}{"interval": interval},
	}
	if planCode != "" {
		obj["metadata"] = map[string]interface{}{"plan_code": planCode}
	}
	d, _ := json.Marshal(obj)
	return d
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
	repoWithCustomer := func() *fakeCustomerRepo {
		r := newFakeCustomerRepo()
		_ = r.Save(context.Background(), &entities.Customer{
			OrgID:            orgID,
			StripeCustomerID: customerID,
		})
		return r
	}

	// Pre-seeded webhook repo with already-processed event_id
	repoWithEvent := func(eventID string) *fakeWebhookEventRepo {
		r := newFakeWebhookEventRepo()
		_, _ = r.Save(context.Background(), &entities.WebhookEvent{
			EventID:   eventID,
			EventType: "any",
			Status:    entities.WebhookEventProcessed,
		})
		return r
	}

	tests := []struct {
		name            string
		rawBody         []byte
		sig             string
		gw              *billingmocks.MockStripeGateway
		customerRepo    *fakeCustomerRepo
		webhookRepo     *fakeWebhookEventRepo
		subRepo         *fakeSubscriptionRepo
		assignErr       error // error MockPlanAssigner.Assign should return
		wantErr         error
		expectConverted bool // true → TrialConverter.Convert MUST be called
		assertFn        func(t *testing.T, pa *billingmocks.MockPlanAssigner, whr *fakeWebhookEventRepo)
	}{
		{
			name:    "checkout.session.completed — first delivery assigns plan and converts trial",
			rawBody: []byte(`{}`),
			sig:     "valid-sig",
			gw: &billingmocks.MockStripeGateway{
				ConstructEventResult:       makeEvent("evt_001", "checkout.session.completed", checkoutData(customerID, subID)),
				RetrieveSubscriptionResult: makeSubResult(customerID, subID, priceID, "active", periodEnd),
			},
			customerRepo:    repoWithCustomer(),
			webhookRepo:     newFakeWebhookEventRepo(),
			subRepo:         newFakeSubscriptionRepo(),
			expectConverted: true,
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *fakeWebhookEventRepo) {
				if pa.AssignCalls != 1 {
					t.Errorf("expected 1 Assign call, got %d", pa.AssignCalls)
				}
			},
		},
		{
			name:    "checkout.session.completed — duplicate event is no-op (no converter call)",
			rawBody: []byte(`{}`),
			sig:     "valid-sig",
			gw: &billingmocks.MockStripeGateway{
				ConstructEventResult: makeEvent("evt_001", "checkout.session.completed", checkoutData(customerID, subID)),
			},
			customerRepo: repoWithCustomer(),
			webhookRepo:  repoWithEvent("evt_001"), // already processed
			subRepo:      newFakeSubscriptionRepo(),
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *fakeWebhookEventRepo) {
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
			webhookRepo:  newFakeWebhookEventRepo(),
			subRepo:      newFakeSubscriptionRepo(),
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *fakeWebhookEventRepo) {
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
			subRepo:      newFakeSubscriptionRepo(),
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *fakeWebhookEventRepo) {
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
			webhookRepo:  newFakeWebhookEventRepo(),
			subRepo:      newFakeSubscriptionRepo(),
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *fakeWebhookEventRepo) {
				if pa.AssignCalls != 1 {
					t.Errorf("subscription.updated: expected 1 Assign call, got %d", pa.AssignCalls)
				}
				if pa.LastAssignIn.StripePriceID != priceID {
					t.Errorf("expected priceID %q, got %q", priceID, pa.LastAssignIn.StripePriceID)
				}
			},
		},
		{
			name:    "customer.subscription.deleted — deactivates plan (no replacement)",
			rawBody: []byte(`{}`),
			sig:     "valid-sig",
			gw: &billingmocks.MockStripeGateway{
				ConstructEventResult: makeEvent("evt_004", "customer.subscription.deleted",
					subData(customerID, subID, priceID, "canceled", periodEnd)),
			},
			customerRepo: repoWithCustomer(),
			webhookRepo:  newFakeWebhookEventRepo(),
			subRepo:      newFakeSubscriptionRepo(),
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *fakeWebhookEventRepo) {
				if pa.DeactivateCalls != 1 {
					t.Errorf("subscription.deleted: expected 1 Deactivate call, got %d", pa.DeactivateCalls)
				}
				if pa.AssignCalls != 0 {
					t.Errorf("subscription.deleted: expected 0 Assign calls, got %d", pa.AssignCalls)
				}
				if pa.LastDeactivateCustID != customerID {
					t.Errorf("expected Deactivate with customer %q, got %q", customerID, pa.LastDeactivateCustID)
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
			webhookRepo:  newFakeWebhookEventRepo(),
			subRepo:      newFakeSubscriptionRepo(),
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *fakeWebhookEventRepo) {
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
			webhookRepo:  newFakeWebhookEventRepo(),
			subRepo:      newFakeSubscriptionRepo(),
			// PlanAssigner returns ErrPlanNotFound for unknown price_id.
			// Handler must treat this as no-op (nil return) per FR5.
			assignErr: services.ErrPlanNotFound,
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *fakeWebhookEventRepo) {
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
			customerRepo: newFakeCustomerRepo(),
			webhookRepo:  newFakeWebhookEventRepo(),
			subRepo:      newFakeSubscriptionRepo(),
			wantErr:      ErrInvalidSignature,
		},
		{
			name:    "missing signature header — returns ErrInvalidSignature",
			rawBody: []byte(`{}`),
			sig:     "", // empty = missing
			gw:      &billingmocks.MockStripeGateway{},
			customerRepo: newFakeCustomerRepo(),
			webhookRepo:  newFakeWebhookEventRepo(),
			subRepo:      newFakeSubscriptionRepo(),
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
			webhookRepo:  newFakeWebhookEventRepo(),
			subRepo:      newFakeSubscriptionRepo(),
			assertFn: func(t *testing.T, pa *billingmocks.MockPlanAssigner, _ *fakeWebhookEventRepo) {
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
			tc := &billingmocks.MockTrialConverter{}
			h := NewHandler(tt.gw, secret, pa, tt.customerRepo, tt.webhookRepo, tt.subRepo).WithTrialConverter(tc)

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
			if tt.expectConverted {
				if tc.ConvertCalls != 1 {
					t.Errorf("expected 1 TrialConverter.Convert call, got %d", tc.ConvertCalls)
				}
				if tc.LastOrgID == uuid.Nil {
					t.Errorf("expected non-Nil orgID passed to TrialConverter, got Nil")
				}
			} else if tc.ConvertCalls != 0 {
				t.Errorf("expected 0 TrialConverter.Convert calls, got %d", tc.ConvertCalls)
			}
			if tt.assertFn != nil {
				tt.assertFn(t, pa, tt.webhookRepo)
			}
		})
	}
}

// TestHandleWebhook_OrphanCustomer_IsDeferred verifies the critical defense-by-
// sync contract: when the Stripe customer has no local org yet, the handler
// MUST mark the event as 'deferred' (NOT 'failed'), persist the raw payload,
// and return nil so the HTTP layer responds with 200 OK. Stripe must not retry.
func TestHandleWebhook_OrphanCustomer_IsDeferred(t *testing.T) {
	const (
		secret     = "whsec_test"
		customerID = "cus_orphan_xyz"
		subID      = "sub_orphan"
		priceID    = "price_orphan"
	)
	eventID := "evt_orphan_001"
	rawBody := []byte(`{"data":{"object":{"customer":"` + customerID + `","subscription":"` + subID + `"}}}`)

	gw := &billingmocks.MockStripeGateway{
		ConstructEventResult:       makeEvent(eventID, "checkout.session.completed", checkoutData(customerID, subID)),
		RetrieveSubscriptionResult: makeSubResult(customerID, subID, priceID, "active", time.Now().Add(30*24*time.Hour).Unix()),
	}
	customerRepo := newFakeCustomerRepo() // empty — no org linked to customer
	webhookRepo := newFakeWebhookEventRepo()
	subRepo := newFakeSubscriptionRepo()
	pa := &billingmocks.MockPlanAssigner{AssignErr: services.ErrOrphanCustomer}

	h := NewHandler(gw, secret, pa, customerRepo, webhookRepo, subRepo)

	err := h.Handle(context.Background(), rawBody, "valid-sig")
	if err != nil {
		t.Fatalf("orphan customer must return nil to HTTP layer, got: %v", err)
	}

	stored, ok := webhookRepo.events[eventID]
	if !ok {
		t.Fatalf("event %q not persisted", eventID)
	}
	if stored.Status != entities.WebhookEventDeferred {
		t.Errorf("status: want %q, got %q", entities.WebhookEventDeferred, stored.Status)
	}
	if len(stored.RawPayload) == 0 {
		t.Error("raw payload must be persisted so ReconcileFromStripe can replay it")
	}
}

func TestHandleWebhook_PricingChanged(t *testing.T) {
	const secret = "whsec_test"

	newHandler := func(gw *billingmocks.MockStripeGateway, pr *fakePlanRepo) *Handler {
		pa := &billingmocks.MockPlanAssigner{}
		return NewHandler(gw, secret, pa,
			newFakeCustomerRepo(), newFakeWebhookEventRepo(), newFakeSubscriptionRepo()).
			WithPlanRepository(pr)
	}

	t.Run("price.updated monthly writes cached amount", func(t *testing.T) {
		pr := &fakePlanRepo{}
		gw := &billingmocks.MockStripeGateway{
			ConstructEventResult: makeEvent("evt_p1", "price.updated",
				priceData("price_m_pro", "prod_pro", 2000, "usd", "month", "pro")),
		}
		if err := newHandler(gw, pr).Handle(context.Background(), []byte(`{}`), "sig"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pr.calls != 1 {
			t.Fatalf("expected 1 upsert, got %d", pr.calls)
		}
		if pr.lastCode != "pro" || pr.lastCycle != "monthly" || pr.lastPrice != "price_m_pro" || pr.lastCents != 2000 || pr.lastCur != "usd" {
			t.Errorf("upsert args = (%q,%q,%q,%d,%q)", pr.lastCode, pr.lastCycle, pr.lastPrice, pr.lastCents, pr.lastCur)
		}
	})

	t.Run("price.created yearly maps interval year -> yearly", func(t *testing.T) {
		pr := &fakePlanRepo{}
		gw := &billingmocks.MockStripeGateway{
			ConstructEventResult: makeEvent("evt_p2", "price.created",
				priceData("price_y_pro", "prod_pro", 19200, "usd", "year", "pro")),
		}
		if err := newHandler(gw, pr).Handle(context.Background(), []byte(`{}`), "sig"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pr.lastCycle != "yearly" || pr.lastCents != 19200 {
			t.Errorf("expected yearly/19200, got %q/%d", pr.lastCycle, pr.lastCents)
		}
	})

	t.Run("missing plan_code is silent no-op", func(t *testing.T) {
		pr := &fakePlanRepo{}
		gw := &billingmocks.MockStripeGateway{
			ConstructEventResult: makeEvent("evt_p3", "price.updated",
				priceData("price_x", "prod_x", 500, "usd", "month", "")),
		}
		if err := newHandler(gw, pr).Handle(context.Background(), []byte(`{}`), "sig"); err != nil {
			t.Fatalf("expected nil (ack), got %v", err)
		}
		if pr.calls != 0 {
			t.Errorf("expected 0 upserts for missing plan_code, got %d", pr.calls)
		}
	})

	t.Run("non-recurring price is skipped", func(t *testing.T) {
		pr := &fakePlanRepo{}
		obj, _ := json.Marshal(map[string]interface{}{
			"id": "price_once", "product": "prod_x", "unit_amount": 500,
			"currency": "usd", "active": true,
			"metadata": map[string]interface{}{"plan_code": "pro"},
		})
		gw := &billingmocks.MockStripeGateway{
			ConstructEventResult: makeEvent("evt_p4", "price.updated", obj),
		}
		if err := newHandler(gw, pr).Handle(context.Background(), []byte(`{}`), "sig"); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
		if pr.calls != 0 {
			t.Errorf("expected 0 upserts for non-recurring price, got %d", pr.calls)
		}
	})

	t.Run("unknown plan_code (ErrPlanNotFound) is acked, not failed", func(t *testing.T) {
		pr := &fakePlanRepo{knownCode: "pro"}
		gw := &billingmocks.MockStripeGateway{
			ConstructEventResult: makeEvent("evt_p5", "price.updated",
				priceData("price_m_ghost", "prod_ghost", 700, "usd", "month", "ghost")),
		}
		if err := newHandler(gw, pr).Handle(context.Background(), []byte(`{}`), "sig"); err != nil {
			t.Errorf("ErrPlanNotFound must be acked as nil, got %v", err)
		}
	})

	t.Run("nil plan repo (not wired) is a no-op", func(t *testing.T) {
		gw := &billingmocks.MockStripeGateway{
			ConstructEventResult: makeEvent("evt_p6", "price.updated",
				priceData("price_m_pro", "prod_pro", 2000, "usd", "month", "pro")),
		}
		pa := &billingmocks.MockPlanAssigner{}
		h := NewHandler(gw, secret, pa, newFakeCustomerRepo(), newFakeWebhookEventRepo(), newFakeSubscriptionRepo())
		if err := h.Handle(context.Background(), []byte(`{}`), "sig"); err != nil {
			t.Errorf("nil plan repo must be a no-op, got %v", err)
		}
	})
}
