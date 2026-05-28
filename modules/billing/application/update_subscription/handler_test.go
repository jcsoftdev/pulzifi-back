package updatesubscription

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	billingmocks "github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services/mocks"
)

type fakePlanRepo struct{ priceID string; err error }

func (r *fakePlanRepo) FindStripePriceID(_ context.Context, _, _ string) (string, error) {
	return r.priceID, r.err
}

type fakeSubRepo struct {
	mu  sync.Mutex
	sub *entities.Subscription
}

func (r *fakeSubRepo) FindByOrgID(_ context.Context, _ uuid.UUID) (*entities.Subscription, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sub == nil {
		return nil, nil
	}
	cp := *r.sub
	return &cp, nil
}

func (r *fakeSubRepo) FindByStripeSubscriptionID(_ context.Context, _ string) (*entities.Subscription, error) {
	return nil, nil
}
func (r *fakeSubRepo) Save(_ context.Context, _ *entities.Subscription) error   { return nil }
func (r *fakeSubRepo) Update(_ context.Context, _ *entities.Subscription) error { return nil }

var _ repositories.SubscriptionRepository = (*fakeSubRepo)(nil)

type fakeUsageReader struct{ snap services.UsageSnapshot }

func (r *fakeUsageReader) ReadActiveUsage(_ context.Context, _ uuid.UUID) (services.UsageSnapshot, error) {
	return r.snap, nil
}

func TestUpdateSubscription_Preview_ComputesUsageBreakdown(t *testing.T) {
	orgID := uuid.New()
	gw := &billingmocks.MockStripeGateway{
		RetrievePriceAmountByID: map[string]int64{
			"price_starter_monthly": 2700, // current
			"price_pro_monthly":     6200, // new
		},
		RetrievePriceCurrency:           "usd",
		RetrieveSubscriptionResult: services.StripeSubscription{
			ID: "sub_existing", PriceID: "price_starter_monthly", Status: "active",
		},
		RetrieveCustomerBalanceCents:    -3500, // $35 credit
		RetrieveCustomerBalanceCurrency: "usd",
	}
	plan := &fakePlanRepo{priceID: "price_pro_monthly"}
	sub := &fakeSubRepo{sub: &entities.Subscription{
		OrgID:                orgID,
		StripeSubscriptionID: "sub_existing",
		StripeCustomerID:     "cus_test",
	}}
	usage := &fakeUsageReader{snap: services.UsageSnapshot{ChecksUsed: 0, ChecksAllowed: 500}}

	h := NewHandler(gw, plan, sub, usage, &billingmocks.MockPlanAssigner{})
	resp, err := h.Handle(context.Background(), Request{
		OrgID:        orgID.String(),
		PlanID:       "pro",
		BillingCycle: "monthly",
		Preview:      true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Usage credit = 2700 × (500/500) = 2700 ($27 starter refunded fully)
	if resp.UsageCreditCents != 2700 {
		t.Errorf("usage credit: want 2700, got %d", resp.UsageCreditCents)
	}
	if resp.AccountCreditCents != 3500 {
		t.Errorf("account credit: want 3500, got %d", resp.AccountCreditCents)
	}
	if resp.NewPlanChargeCents != 6200 {
		t.Errorf("new plan charge: want 6200, got %d", resp.NewPlanChargeCents)
	}
	// 6200 - 2700 - 3500 = 0
	if resp.AmountDueCents != 0 {
		t.Errorf("amount due: want 0, got %d", resp.AmountDueCents)
	}
	if gw.CreateRefundCreditInvoiceItemCalls != 0 {
		t.Errorf("preview must NOT create invoice item")
	}
	if gw.UpdateSubscriptionAnchorNowCalls != 0 {
		t.Errorf("preview must NOT mutate subscription")
	}
}

func TestUpdateSubscription_Preview_PartialUsage(t *testing.T) {
	orgID := uuid.New()
	gw := &billingmocks.MockStripeGateway{
		RetrievePriceAmountByID: map[string]int64{
			"price_starter_monthly": 2700,
			"price_pro_monthly":     6200,
		},
		RetrievePriceCurrency:    "usd",
		RetrieveCustomerBalanceCents: 0,
		RetrieveSubscriptionResult: services.StripeSubscription{
			ID: "sub_existing", PriceID: "price_starter_monthly", Status: "active",
		},
	}
	plan := &fakePlanRepo{priceID: "price_pro_monthly"}
	sub := &fakeSubRepo{sub: &entities.Subscription{
		OrgID:                orgID,
		StripeSubscriptionID: "sub_existing",
		StripeCustomerID:     "cus_test",
	}}
	// 250/500 used → 250 remaining → ratio 0.5 → credit = 2700 × 0.5 = 1350
	usage := &fakeUsageReader{snap: services.UsageSnapshot{ChecksUsed: 250, ChecksAllowed: 500}}

	h := NewHandler(gw, plan, sub, usage, &billingmocks.MockPlanAssigner{})
	resp, err := h.Handle(context.Background(), Request{
		OrgID: orgID.String(), PlanID: "pro", BillingCycle: "monthly", Preview: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.UsageCreditCents != 1350 {
		t.Errorf("usage credit: want 1350, got %d", resp.UsageCreditCents)
	}
	if resp.AmountDueCents != 6200-1350 {
		t.Errorf("amount due: want %d, got %d", 6200-1350, resp.AmountDueCents)
	}
}

func TestUpdateSubscription_Apply_CreatesRefundAndUpdatesSub(t *testing.T) {
	orgID := uuid.New()
	gw := &billingmocks.MockStripeGateway{
		RetrievePriceAmountByID: map[string]int64{
			"price_starter_monthly": 2700,
			"price_pro_monthly":     6200,
		},
		RetrievePriceCurrency:        "usd",
		RetrieveCustomerBalanceCents: 0,
		RetrieveSubscriptionResult: services.StripeSubscription{
			ID: "sub_existing", PriceID: "price_starter_monthly", Status: "active",
		},
		UpdateSubscriptionAnchorNowResult: services.StripeSubscription{
			ID: "sub_existing", Status: "active", PriceID: "price_pro_monthly",
		},
	}
	plan := &fakePlanRepo{priceID: "price_pro_monthly"}
	sub := &fakeSubRepo{sub: &entities.Subscription{
		OrgID:                orgID,
		StripeSubscriptionID: "sub_existing",
		StripeCustomerID:     "cus_test",
	}}
	usage := &fakeUsageReader{snap: services.UsageSnapshot{ChecksUsed: 0, ChecksAllowed: 500}}

	h := NewHandler(gw, plan, sub, usage, &billingmocks.MockPlanAssigner{})
	_, err := h.Handle(context.Background(), Request{
		OrgID: orgID.String(), PlanID: "pro", BillingCycle: "monthly",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gw.CreateRefundCreditInvoiceItemCalls != 1 {
		t.Errorf("want 1 CreateRefundCreditInvoiceItem call, got %d", gw.CreateRefundCreditInvoiceItemCalls)
	}
	if gw.LastRefundCreditAmount != 2700 {
		t.Errorf("refund amount: want 2700, got %d", gw.LastRefundCreditAmount)
	}
	if gw.UpdateSubscriptionAnchorNowCalls != 1 {
		t.Errorf("want 1 UpdateSubscriptionAnchorNow call, got %d", gw.UpdateSubscriptionAnchorNowCalls)
	}
}

func TestUpdateSubscription_Apply_SkipsRefundWhenAllChecksUsed(t *testing.T) {
	orgID := uuid.New()
	gw := &billingmocks.MockStripeGateway{
		RetrievePriceAmountByID: map[string]int64{
			"price_starter_monthly": 2700,
			"price_pro_monthly":     6200,
		},
		RetrievePriceCurrency: "usd",
		UpdateSubscriptionAnchorNowResult: services.StripeSubscription{
			ID: "sub_existing", Status: "active", PriceID: "price_pro_monthly",
		},
	}
	plan := &fakePlanRepo{priceID: "price_pro_monthly"}
	sub := &fakeSubRepo{sub: &entities.Subscription{
		OrgID:                orgID,
		StripeSubscriptionID: "sub_existing",
		StripeCustomerID:     "cus_test",
	}}
	// All 500 used → 0 remaining → no credit
	usage := &fakeUsageReader{snap: services.UsageSnapshot{ChecksUsed: 500, ChecksAllowed: 500}}

	h := NewHandler(gw, plan, sub, usage, &billingmocks.MockPlanAssigner{})
	_, err := h.Handle(context.Background(), Request{
		OrgID: orgID.String(), PlanID: "pro", BillingCycle: "monthly",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gw.CreateRefundCreditInvoiceItemCalls != 0 {
		t.Errorf("must NOT create refund when no remaining checks, got %d calls", gw.CreateRefundCreditInvoiceItemCalls)
	}
}

func TestUpdateSubscription_InvalidCycle_Rejected(t *testing.T) {
	h := NewHandler(&billingmocks.MockStripeGateway{}, &fakePlanRepo{}, &fakeSubRepo{}, &fakeUsageReader{}, &billingmocks.MockPlanAssigner{})
	_, err := h.Handle(context.Background(), Request{
		OrgID: uuid.New().String(), PlanID: "pro", BillingCycle: "daily",
	})
	if err != ErrInvalidBillingCycle {
		t.Fatalf("expected ErrInvalidBillingCycle, got %v", err)
	}
}

func TestUpdateSubscription_NoActiveSub_Rejected(t *testing.T) {
	h := NewHandler(&billingmocks.MockStripeGateway{}, &fakePlanRepo{priceID: "price_x"}, &fakeSubRepo{}, &fakeUsageReader{}, &billingmocks.MockPlanAssigner{})
	_, err := h.Handle(context.Background(), Request{
		OrgID: uuid.New().String(), PlanID: "pro", BillingCycle: "monthly",
	})
	if err != ErrNoActiveSubscription {
		t.Fatalf("expected ErrNoActiveSubscription, got %v", err)
	}
}
