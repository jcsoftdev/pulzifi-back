package orgwiring

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	billingservices "github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services/mocks"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
)

// ── in-memory fake for planReader ────────────────────────────────────────────

type fakePlanRow struct {
	subscriptionID string
	billingStatus  string
	amount         int64
}

type fakePlanDB struct {
	rows map[uuid.UUID]*fakePlanRow
}

func (f *fakePlanDB) readActivePlan(_ context.Context, orgID uuid.UUID) (*planRow, error) {
	if f.rows == nil {
		return nil, sql.ErrNoRows
	}
	r, ok := f.rows[orgID]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return &planRow{
		subscriptionID: r.subscriptionID,
		billingStatus:  r.billingStatus,
		amount:         r.amount,
	}, nil
}

func fakePlanDBWith(orgID uuid.UUID, row *fakePlanRow) *fakePlanDB {
	if row == nil {
		return &fakePlanDB{rows: nil}
	}
	return &fakePlanDB{rows: map[uuid.UUID]*fakePlanRow{orgID: row}}
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestBillingCancelAdapter_NoPlanRow(t *testing.T) {
	db := &fakePlanDB{rows: nil}
	gw := &mocks.MockStripeGateway{}
	adapter := NewBillingCancelAdapterForTest(db, gw)

	err := adapter.CancelForOrg(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("expected nil when no plan row, got %v", err)
	}
	if gw.CancelSubscriptionNowCalls != 0 {
		t.Errorf("gateway should not be called when no plan row")
	}
}

func TestBillingCancelAdapter_TrialSubCancelSucceeds(t *testing.T) {
	orgID := uuid.New()
	db := fakePlanDBWith(orgID, &fakePlanRow{subscriptionID: "sub_trial", billingStatus: "trialing", amount: 0})
	gw := &mocks.MockStripeGateway{
		CancelSubscriptionNowResult: billingservices.StripeSubscription{ID: "sub_trial", Status: "canceled"},
	}
	adapter := NewBillingCancelAdapterForTest(db, gw)

	err := adapter.CancelForOrg(context.Background(), orgID)
	if err != nil {
		t.Fatalf("trialing cancel success → expected nil, got %v", err)
	}
}

func TestBillingCancelAdapter_TrialSubCancelFails(t *testing.T) {
	orgID := uuid.New()
	db := fakePlanDBWith(orgID, &fakePlanRow{subscriptionID: "sub_trial", billingStatus: "trialing", amount: 0})
	gw := &mocks.MockStripeGateway{
		CancelSubscriptionNowErr: errors.New("stripe error"),
	}
	adapter := NewBillingCancelAdapterForTest(db, gw)

	err := adapter.CancelForOrg(context.Background(), orgID)
	if err != nil {
		t.Fatalf("trialing cancel failure → expected nil (non-abort), got %v", err)
	}
}

func TestBillingCancelAdapter_PaidActiveCancelSucceeds(t *testing.T) {
	orgID := uuid.New()
	db := fakePlanDBWith(orgID, &fakePlanRow{subscriptionID: "sub_paid", billingStatus: "active", amount: 1500})
	gw := &mocks.MockStripeGateway{
		CancelSubscriptionNowResult: billingservices.StripeSubscription{ID: "sub_paid", Status: "canceled"},
	}
	adapter := NewBillingCancelAdapterForTest(db, gw)

	err := adapter.CancelForOrg(context.Background(), orgID)
	if err != nil {
		t.Fatalf("paid active cancel success → expected nil, got %v", err)
	}
}

func TestBillingCancelAdapter_PaidActiveCancelFails(t *testing.T) {
	orgID := uuid.New()
	db := fakePlanDBWith(orgID, &fakePlanRow{subscriptionID: "sub_paid", billingStatus: "active", amount: 1500})
	gw := &mocks.MockStripeGateway{
		CancelSubscriptionNowErr: errors.New("rate limit exceeded"),
	}
	adapter := NewBillingCancelAdapterForTest(db, gw)

	err := adapter.CancelForOrg(context.Background(), orgID)
	if !errors.Is(err, orgservices.ErrBillingActive) {
		t.Fatalf("paid active cancel failure → expected ErrBillingActive, got %v", err)
	}
}

func TestBillingCancelAdapter_PastDueCancelFails(t *testing.T) {
	orgID := uuid.New()
	db := fakePlanDBWith(orgID, &fakePlanRow{subscriptionID: "sub_pastdue", billingStatus: "past_due", amount: 999})
	gw := &mocks.MockStripeGateway{
		CancelSubscriptionNowErr: errors.New("network timeout"),
	}
	adapter := NewBillingCancelAdapterForTest(db, gw)

	err := adapter.CancelForOrg(context.Background(), orgID)
	if !errors.Is(err, orgservices.ErrBillingActive) {
		t.Fatalf("past_due cancel failure → expected ErrBillingActive, got %v", err)
	}
}

func TestNopBillingCanceller_AlwaysNil(t *testing.T) {
	nop := NopBillingCanceller()
	err := nop.CancelForOrg(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("nop canceller must return nil, got %v", err)
	}
}
