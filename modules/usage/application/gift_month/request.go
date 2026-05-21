package giftmonth

import "github.com/google/uuid"

// Request carries the organization ID for the gift.
type Request struct {
	OrgID uuid.UUID
}
