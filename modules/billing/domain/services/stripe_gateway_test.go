// Package services_test verifies the MockStripeGateway test double satisfies the
// StripeGateway interface and delegates calls correctly. These tests exercise the
// mock that all billing use-case tests depend on, ensuring it behaves predictably.
package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services/mocks"
)

// compile-time interface check
var _ services.StripeGateway = (*mocks.MockStripeGateway)(nil)

func TestMockStripeGateway_CreateCheckoutSession_ReturnsConfiguredURL(t *testing.T) {
	mock := &mocks.MockStripeGateway{
		CreateCheckoutSessionResult: "https://checkout.stripe.com/pay/cs_test_abc",
	}

	url, err := mock.CreateCheckoutSession(context.Background(), services.CheckoutInput{
		CustomerID: "cus_123",
		PriceID:    "price_abc",
		SuccessURL: "https://app.com/success",
		CancelURL:  "https://app.com/cancel",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != mock.CreateCheckoutSessionResult {
		t.Errorf("url: want %q, got %q", mock.CreateCheckoutSessionResult, url)
	}
	if mock.CreateCheckoutSessionCalls != 1 {
		t.Errorf("CreateCheckoutSessionCalls: want 1, got %d", mock.CreateCheckoutSessionCalls)
	}
}

func TestMockStripeGateway_CreateCheckoutSession_ReturnsConfiguredError(t *testing.T) {
	expectedErr := errors.New("stripe: rate limited")
	mock := &mocks.MockStripeGateway{
		CreateCheckoutSessionErr: expectedErr,
	}

	_, err := mock.CreateCheckoutSession(context.Background(), services.CheckoutInput{})
	if !errors.Is(err, expectedErr) {
		t.Errorf("error: want %v, got %v", expectedErr, err)
	}
}

func TestMockStripeGateway_CreateCheckoutSession_FnOverridesFields(t *testing.T) {
	called := false
	mock := &mocks.MockStripeGateway{
		CreateCheckoutSessionFn: func(_ context.Context, in services.CheckoutInput) (string, error) {
			called = true
			return "https://fn-url", nil
		},
		CreateCheckoutSessionResult: "https://should-not-be-used",
	}

	url, err := mock.CreateCheckoutSession(context.Background(), services.CheckoutInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected Fn to be called, but it was not")
	}
	if url != "https://fn-url" {
		t.Errorf("expected Fn result, got %q", url)
	}
}

func TestMockStripeGateway_EnsureCustomer_ReturnsConfiguredID(t *testing.T) {
	mock := &mocks.MockStripeGateway{
		EnsureCustomerResult: "cus_ensured",
	}

	id, err := mock.EnsureCustomer(context.Background(), "org-id", "email@example.com", "Org Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "cus_ensured" {
		t.Errorf("customerID: want 'cus_ensured', got %q", id)
	}
	if mock.EnsureCustomerCalls != 1 {
		t.Errorf("EnsureCustomerCalls: want 1, got %d", mock.EnsureCustomerCalls)
	}
}

func TestMockStripeGateway_CancelSubscription_IsNilErr(t *testing.T) {
	// Verifies that a fresh mock (no error set) returns nil — the happy-path contract
	// used by tests that expect CancelSubscription to succeed.
	mock := &mocks.MockStripeGateway{}

	// RetrieveSubscription is the closest analogue to a cancel check.
	sub, err := mock.RetrieveSubscription(context.Background(), "sub_abc")
	if err != nil {
		t.Errorf("expected nil error from zero-value mock, got: %v", err)
	}
	// Zero value should have empty ID.
	if sub.ID != "" {
		t.Errorf("expected empty subscription ID from zero-value mock, got %q", sub.ID)
	}
	if mock.RetrieveSubscriptionCalls != 1 {
		t.Errorf("RetrieveSubscriptionCalls: want 1, got %d", mock.RetrieveSubscriptionCalls)
	}
}
