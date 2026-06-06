package giftmonth

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	billingmocks "github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services/mocks"
)

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

type fakePlanRepo struct {
	byCycle map[string]string // keyed by "<code>/<cycle>"
	err     error
}

func (r *fakePlanRepo) FindStripePriceID(_ context.Context, code, cycle string) (string, error) {
	if r.err != nil {
		return "", r.err
	}
	return r.byCycle[code+"/"+cycle], nil
}

func (r *fakePlanRepo) UpsertStripePricing(_ context.Context, _, _, _ string, _ int64, _ string) error {
	return nil
}

func activeSub(orgID uuid.UUID, cycle string) *fakeSubRepo {
	return &fakeSubRepo{sub: &entities.Subscription{
		OrgID:                orgID,
		StripeSubscriptionID: "sub_active",
		StripeCustomerID:     "cus_active",
		PlanCode:             "starter",
		BillingCycle:         cycle,
	}}
}

func TestGift_HigherTier_UsesSchedule(t *testing.T) {
	orgID := uuid.New()
	gw := &billingmocks.MockStripeGateway{
		RetrievePriceAmountByID:  map[string]int64{"price_pro_monthly": 6200},
		RetrievePriceCurrency:    "usd",
		RetrieveSubscriptionResult: services.StripeSubscription{ID: "sub_active", PriceID: "price_starter_current"},
		GiftPlanScheduleRevertAt: 1730000000,
	}
	plan := &fakePlanRepo{byCycle: map[string]string{"pro/monthly": "price_pro_monthly"}}
	// Org on Starter, admin gifts PRO (higher tier) → free Pro month via schedule.
	sub := activeSub(orgID, "monthly")

	h := NewHandler(gw, sub, plan, &billingmocks.MockPlanAssigner{})
	resp, err := h.Handle(context.Background(), orgID, "pro", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gw.GiftPlanScheduleCalls != 1 {
		t.Errorf("expected 1 schedule call, got %d", gw.GiftPlanScheduleCalls)
	}
	if gw.CreditCustomerBalanceCalls != 0 {
		t.Errorf("expected 0 balance credits on upgrade gift, got %d", gw.CreditCustomerBalanceCalls)
	}
	if gw.LastGiftScheduleGiftPriceID != "price_pro_monthly" {
		t.Errorf("schedule gift price wrong: %q", gw.LastGiftScheduleGiftPriceID)
	}
	if resp.Mode != ModePlanGift || resp.GiftPlanCode != "pro" || resp.AmountCents != 6200 || resp.RevertAt != 1730000000 {
		t.Errorf("response wrong: %+v", resp)
	}
}

func TestGift_SameOrLowerTier_BanksBalance(t *testing.T) {
	orgID := uuid.New()
	gw := &billingmocks.MockStripeGateway{
		RetrievePriceAmountByID: map[string]int64{"price_starter_monthly": 2700},
		RetrievePriceCurrency:   "usd",
	}
	plan := &fakePlanRepo{byCycle: map[string]string{"starter/monthly": "price_starter_monthly"}}
	// Org on Starter, gift Starter (same tier) → banked balance credit.
	sub := activeSub(orgID, "monthly")

	h := NewHandler(gw, sub, plan, &billingmocks.MockPlanAssigner{})
	resp, err := h.Handle(context.Background(), orgID, "starter", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gw.GiftPlanScheduleCalls != 0 {
		t.Errorf("expected 0 schedule calls on same-tier gift, got %d", gw.GiftPlanScheduleCalls)
	}
	if gw.LastCreditAmountCents != 2700 {
		t.Errorf("credit amount: want 2700 (Starter monthly), got %d", gw.LastCreditAmountCents)
	}
	if resp.Mode != ModeBalanceCredit {
		t.Errorf("mode: want balance_credit, got %q", resp.Mode)
	}
}

func TestGift_NewOrg_CreatesFreeTrialSub(t *testing.T) {
	orgID := uuid.New()
	gw := &billingmocks.MockStripeGateway{
		RetrievePriceAmountByID:      map[string]int64{"price_pro_monthly": 6200},
		RetrievePriceCurrency:        "usd",
		EnsureCustomerResult:         "cus_new",
		CreateGiftSubscriptionResult: services.StripeSubscription{ID: "sub_gift", CurrentPeriodEnd: 1730000000},
	}
	plan := &fakePlanRepo{byCycle: map[string]string{"pro/monthly": "price_pro_monthly"}}
	// No active subscription (new / trial org).
	h := NewHandler(gw, &fakeSubRepo{}, plan, &billingmocks.MockPlanAssigner{})

	resp, err := h.Handle(context.Background(), orgID, "pro", "owner@example.com", "New Org")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gw.CreateGiftSubscriptionCalls != 1 {
		t.Errorf("expected 1 gift subscription creation, got %d", gw.CreateGiftSubscriptionCalls)
	}
	if gw.LastGiftSubCustomerID != "cus_new" || gw.LastGiftSubPriceID != "price_pro_monthly" {
		t.Errorf("gift sub args wrong: cust=%q price=%q", gw.LastGiftSubCustomerID, gw.LastGiftSubPriceID)
	}
	if resp.Mode != ModePlanGift || resp.RevertAt != 1730000000 {
		t.Errorf("response wrong: %+v", resp)
	}
}

func TestGift_PlanNotFound(t *testing.T) {
	orgID := uuid.New()
	plan := &fakePlanRepo{err: repositories.ErrPlanNotFound}
	h := NewHandler(&billingmocks.MockStripeGateway{}, activeSub(orgID, "monthly"), plan, &billingmocks.MockPlanAssigner{})
	_, err := h.Handle(context.Background(), orgID, "ghost", "", "")
	if !errors.Is(err, ErrPlanNotFound) {
		t.Fatalf("want ErrPlanNotFound, got %v", err)
	}
}

func TestGift_BalanceCreditError_Propagates(t *testing.T) {
	orgID := uuid.New()
	gw := &billingmocks.MockStripeGateway{
		RetrievePriceAmountByID:  map[string]int64{"price_starter_monthly": 2700},
		RetrievePriceCurrency:    "usd",
		CreditCustomerBalanceErr: errors.New("stripe down"),
	}
	plan := &fakePlanRepo{byCycle: map[string]string{"starter/monthly": "price_starter_monthly"}}
	// Same-tier gift → balance path → error must propagate.
	h := NewHandler(gw, activeSub(orgID, "monthly"), plan, &billingmocks.MockPlanAssigner{})
	_, err := h.Handle(context.Background(), orgID, "starter", "", "")
	if err == nil {
		t.Fatal("expected balance credit error to propagate")
	}
}
