package services

import (
	"context"

	"github.com/google/uuid"
)

// TrialConverter is invoked after a successful trial-to-paid conversion to
// mark the trial row as converted and lift any user-level trial gates.
//
// Implementations live in cmd/wiring/billing/ to keep the billing module
// decoupled from auth and organization infrastructure packages.
type TrialConverter interface {
	// Convert marks all active trial organization_plans rows for the org as
	// converted (sets converted_at = now()) and flips any 'trial_expired'
	// users in that org back to 'approved'.
	//
	// It is safe to call this multiple times: re-running on an already
	// converted org is a no-op.
	Convert(ctx context.Context, orgID uuid.UUID) error
}
