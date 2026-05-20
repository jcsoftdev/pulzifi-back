package entities

import "fmt"

// BillingStatus represents the subscription billing state as reported by Stripe.
type BillingStatus string

const (
	BillingActive     BillingStatus = "active"
	BillingPastDue    BillingStatus = "past_due"
	BillingCanceled   BillingStatus = "canceled"
	BillingIncomplete BillingStatus = "incomplete"
	BillingTrialing   BillingStatus = "trialing"
)

// String returns the string representation of the BillingStatus.
func (s BillingStatus) String() string {
	return string(s)
}

// IsValid reports whether the status is one of the known Stripe billing states.
func (s BillingStatus) IsValid() bool {
	switch s {
	case BillingActive, BillingPastDue, BillingCanceled, BillingIncomplete, BillingTrialing:
		return true
	}
	return false
}

// BillingStatusFromString parses a raw string into a BillingStatus.
// Returns an error if the value is not a recognised status.
func BillingStatusFromString(s string) (BillingStatus, error) {
	bs := BillingStatus(s)
	if !bs.IsValid() {
		return "", fmt.Errorf("billing: unknown billing status %q", s)
	}
	return bs, nil
}
