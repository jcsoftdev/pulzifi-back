package managecoupons

import (
	"context"
	"errors"
	"testing"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	billingmocks "github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services/mocks"
)

type fakePlanRepo struct {
	priceID string
	err     error
}

func (r *fakePlanRepo) FindStripePriceID(_ context.Context, _, _ string) (string, error) {
	return r.priceID, r.err
}

func TestCreate_FirstMonthOneDollar_ComputesAmountOff(t *testing.T) {
	gw := &billingmocks.MockStripeGateway{
		RetrievePriceAmountByID: map[string]int64{"price_pro_monthly": 6200},
		RetrievePriceCurrency:   "usd",
		CreateCouponResult:      "coupon_123",
		CreatePromotionCodeResult: services.PromotionCode{
			ID: "promo_123", Code: "TRY1PRO", CouponID: "coupon_123",
			AmountOffCents: 6100, Currency: "usd", Active: true,
		},
	}
	plan := &fakePlanRepo{priceID: "price_pro_monthly"}
	h := NewHandler(gw, plan, "https://app.pulzifi.com/")

	resp, err := h.Create(context.Background(), CreateRequest{
		PlanCode: "pro", BillingCycle: "monthly", Code: "try1pro",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 6200 - 100 = 6100 amount_off so first cycle = $1
	if gw.LastCouponAmount != 6100 {
		t.Errorf("amount_off: want 6100, got %d", gw.LastCouponAmount)
	}
	if resp.Code != "TRY1PRO" {
		t.Errorf("code: want TRY1PRO, got %q", resp.Code)
	}
	if resp.ApplyURL != "https://app.pulzifi.com/settings/billing?promo=TRY1PRO" {
		t.Errorf("apply_url wrong: %q", resp.ApplyURL)
	}
}

func TestCreate_InvalidCycle(t *testing.T) {
	h := NewHandler(&billingmocks.MockStripeGateway{}, &fakePlanRepo{}, "https://x")
	_, err := h.Create(context.Background(), CreateRequest{PlanCode: "pro", BillingCycle: "daily"})
	if err != ErrInvalidBillingCycle {
		t.Fatalf("want ErrInvalidBillingCycle, got %v", err)
	}
}

func TestCreate_PlanNotFound(t *testing.T) {
	plan := &fakePlanRepo{err: repositories.ErrPlanNotFound}
	h := NewHandler(&billingmocks.MockStripeGateway{}, plan, "https://x")
	_, err := h.Create(context.Background(), CreateRequest{PlanCode: "ghost", BillingCycle: "monthly"})
	if err != ErrPlanNotFound {
		t.Fatalf("want ErrPlanNotFound, got %v", err)
	}
}

func TestCreate_PriceBelowOneDollar_ClampsToZero(t *testing.T) {
	gw := &billingmocks.MockStripeGateway{
		RetrievePriceAmountByID:   map[string]int64{"price_cheap": 50}, // $0.50 < $1
		RetrievePriceCurrency:     "usd",
		CreateCouponResult:        "coupon_x",
		CreatePromotionCodeResult: services.PromotionCode{ID: "promo_x", Code: "CHEAP"},
	}
	plan := &fakePlanRepo{priceID: "price_cheap"}
	h := NewHandler(gw, plan, "https://x")
	_, err := h.Create(context.Background(), CreateRequest{PlanCode: "cheap", BillingCycle: "monthly"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gw.LastCouponAmount != 0 {
		t.Errorf("amount_off must clamp to 0 when price < $1, got %d", gw.LastCouponAmount)
	}
}

func TestList_MapsPromotionCodes(t *testing.T) {
	gw := &billingmocks.MockStripeGateway{
		ListPromotionCodesResult: []services.PromotionCode{
			{ID: "p1", Code: "A", AmountOffCents: 6100, Active: true, PlanCode: "pro", BillingCycle: "monthly"},
			{ID: "p2", Code: "B", AmountOffCents: 2600, Active: false, PlanCode: "starter", BillingCycle: "monthly"},
		},
	}
	h := NewHandler(gw, &fakePlanRepo{}, "https://app")
	list, err := h.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2 codes, got %d", len(list))
	}
	if list[0].ApplyURL != "https://app/settings/billing?promo=A" {
		t.Errorf("apply_url[0] wrong: %q", list[0].ApplyURL)
	}
}

func TestRevoke_CallsDeactivate(t *testing.T) {
	gw := &billingmocks.MockStripeGateway{}
	h := NewHandler(gw, &fakePlanRepo{}, "https://x")
	if err := h.Revoke(context.Background(), "promo_123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gw.DeactivatePromotionCodeCalls != 1 || gw.LastDeactivatedPromoID != "promo_123" {
		t.Errorf("deactivate not called correctly: calls=%d id=%q", gw.DeactivatePromotionCodeCalls, gw.LastDeactivatedPromoID)
	}
}

func TestRevoke_EmptyID(t *testing.T) {
	h := NewHandler(&billingmocks.MockStripeGateway{}, &fakePlanRepo{}, "https://x")
	if err := h.Revoke(context.Background(), ""); err == nil {
		t.Fatal("expected error on empty id")
	}
}

func TestCreate_CouponError_Propagates(t *testing.T) {
	gw := &billingmocks.MockStripeGateway{
		RetrievePriceAmountByID: map[string]int64{"price_pro": 6200},
		RetrievePriceCurrency:   "usd",
		CreateCouponErr:         errors.New("stripe down"),
	}
	plan := &fakePlanRepo{priceID: "price_pro"}
	h := NewHandler(gw, plan, "https://x")
	_, err := h.Create(context.Background(), CreateRequest{PlanCode: "pro", BillingCycle: "monthly"})
	if err == nil {
		t.Fatal("expected coupon creation error to propagate")
	}
}
