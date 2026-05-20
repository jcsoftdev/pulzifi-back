package entities_test

import (
	"testing"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
)

func TestBillingStatusString(t *testing.T) {
	tests := []struct {
		status   entities.BillingStatus
		expected string
	}{
		{entities.BillingActive, "active"},
		{entities.BillingPastDue, "past_due"},
		{entities.BillingCanceled, "canceled"},
		{entities.BillingIncomplete, "incomplete"},
		{entities.BillingTrialing, "trialing"},
	}

	for _, tc := range tests {
		t.Run(string(tc.status), func(t *testing.T) {
			if got := tc.status.String(); got != tc.expected {
				t.Errorf("BillingStatus.String() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestBillingStatusIsValid(t *testing.T) {
	validStatuses := []entities.BillingStatus{
		entities.BillingActive,
		entities.BillingPastDue,
		entities.BillingCanceled,
		entities.BillingIncomplete,
		entities.BillingTrialing,
	}

	for _, s := range validStatuses {
		t.Run("valid_"+string(s), func(t *testing.T) {
			if !s.IsValid() {
				t.Errorf("BillingStatus(%q).IsValid() = false, want true", s)
			}
		})
	}
}

func TestBillingStatusIsValidRejectsUnknown(t *testing.T) {
	unknowns := []entities.BillingStatus{
		"",
		"unknown",
		"ACTIVE",  // case-sensitive
		"cancelled", // wrong spelling
	}

	for _, s := range unknowns {
		t.Run("invalid_"+string(s), func(t *testing.T) {
			if s.IsValid() {
				t.Errorf("BillingStatus(%q).IsValid() = true, want false", s)
			}
		})
	}
}

func TestBillingStatusFromString(t *testing.T) {
	tests := []struct {
		input   string
		want    entities.BillingStatus
		wantErr bool
	}{
		{"active", entities.BillingActive, false},
		{"past_due", entities.BillingPastDue, false},
		{"canceled", entities.BillingCanceled, false},
		{"incomplete", entities.BillingIncomplete, false},
		{"trialing", entities.BillingTrialing, false},
		{"unknown", "", true},
		{"", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got, err := entities.BillingStatusFromString(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("BillingStatusFromString(%q) expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Errorf("BillingStatusFromString(%q) unexpected error: %v", tc.input, err)
				return
			}
			if got != tc.want {
				t.Errorf("BillingStatusFromString(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
