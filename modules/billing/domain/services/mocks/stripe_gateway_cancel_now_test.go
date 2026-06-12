package mocks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services/mocks"
)

// compile-time: MockStripeGateway must satisfy StripeGateway including the new method.
var _ services.StripeGateway = (*mocks.MockStripeGateway)(nil)

func TestMockStripeGateway_CancelSubscriptionNow_ReturnsConfiguredResult(t *testing.T) {
	want := services.StripeSubscription{
		ID:     "sub_cancelled",
		Status: "canceled",
	}
	m := &mocks.MockStripeGateway{
		CancelSubscriptionNowResult: want,
	}

	got, err := m.CancelSubscriptionNow(context.Background(), "sub_cancelled")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ID: want %q, got %q", want.ID, got.ID)
	}
	if got.Status != want.Status {
		t.Errorf("Status: want %q, got %q", want.Status, got.Status)
	}
	if m.CancelSubscriptionNowCalls != 1 {
		t.Errorf("CancelSubscriptionNowCalls: want 1, got %d", m.CancelSubscriptionNowCalls)
	}
}

func TestMockStripeGateway_CancelSubscriptionNow_ReturnsConfiguredError(t *testing.T) {
	wantErr := errors.New("stripe: rate limit exceeded")
	m := &mocks.MockStripeGateway{
		CancelSubscriptionNowErr: wantErr,
	}

	_, err := m.CancelSubscriptionNow(context.Background(), "sub_abc")
	if !errors.Is(err, wantErr) {
		t.Errorf("error: want %v, got %v", wantErr, err)
	}
}

func TestMockStripeGateway_CancelSubscriptionNow_FnOverridesFields(t *testing.T) {
	called := false
	wantSub := services.StripeSubscription{ID: "sub_fn", Status: "canceled"}
	m := &mocks.MockStripeGateway{
		CancelSubscriptionNowFn: func(_ context.Context, subID string) (services.StripeSubscription, error) {
			called = true
			return wantSub, nil
		},
		CancelSubscriptionNowResult: services.StripeSubscription{ID: "sub_should_not_be_used"},
	}

	got, err := m.CancelSubscriptionNow(context.Background(), "sub_fn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected Fn to be called, but it was not")
	}
	if got.ID != wantSub.ID {
		t.Errorf("ID: want %q, got %q", wantSub.ID, got.ID)
	}
}
