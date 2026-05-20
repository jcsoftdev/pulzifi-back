package entities

import "github.com/google/uuid"

// Customer maps an organisation to its Stripe customer record.
type Customer struct {
	OrgID            uuid.UUID
	StripeCustomerID string
}
