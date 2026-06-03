package intwiring

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// stubPlanLookup is a test double for OrgPlanLookup that avoids a real DB.
type stubPlanLookup struct {
	code string
	err  error
}

func (s *stubPlanLookup) PlanCode(_ context.Context, _ uuid.UUID) (string, error) {
	return s.code, s.err
}

// channelEntitlementWithStub is a test-only constructor that accepts the
// stubPlanLookup interface instead of the concrete *OrgPlanLookup.
// This keeps the production type un-polluted while allowing unit tests.
type planCodeer interface {
	PlanCode(ctx context.Context, orgID uuid.UUID) (string, error)
}

type testChannelEntitlement struct {
	plans     planCodeer
	paidPlans []string
}

func (a *testChannelEntitlement) IsPaid(ctx context.Context, orgID uuid.UUID) (bool, error) {
	code, err := a.plans.PlanCode(ctx, orgID)
	if err != nil {
		return false, err
	}
	for _, p := range a.paidPlans {
		if p == code {
			return true, nil
		}
	}
	return false, nil
}

var defaultPaidPlans = []string{"trial", "starter", "pro", "enterprise"}

func TestChannelEntitlement_PaidPlan(t *testing.T) {
	cases := []struct {
		code string
	}{
		{"trial"},
		{"starter"},
		{"pro"},
		{"enterprise"},
	}
	for _, tc := range cases {
		t.Run(tc.code, func(t *testing.T) {
			adapter := &testChannelEntitlement{
				plans:     &stubPlanLookup{code: tc.code},
				paidPlans: defaultPaidPlans,
			}
			paid, err := adapter.IsPaid(context.Background(), uuid.New())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !paid {
				t.Errorf("plan %q should be paid", tc.code)
			}
		})
	}
}

func TestChannelEntitlement_FreePlan_EmptyCode(t *testing.T) {
	adapter := &testChannelEntitlement{
		plans:     &stubPlanLookup{code: ""},
		paidPlans: defaultPaidPlans,
	}
	paid, err := adapter.IsPaid(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if paid {
		t.Error("empty plan code should be free")
	}
}

func TestChannelEntitlement_UnknownPlanCode_Free(t *testing.T) {
	adapter := &testChannelEntitlement{
		plans:     &stubPlanLookup{code: "legacy"},
		paidPlans: defaultPaidPlans,
	}
	paid, err := adapter.IsPaid(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if paid {
		t.Error("unknown plan code should be treated as free")
	}
}

func TestChannelEntitlement_LookupError_Propagated(t *testing.T) {
	adapter := &testChannelEntitlement{
		plans:     &stubPlanLookup{err: context.DeadlineExceeded},
		paidPlans: defaultPaidPlans,
	}
	_, err := adapter.IsPaid(context.Background(), uuid.New())
	if err == nil {
		t.Error("expected error to be propagated")
	}
}
