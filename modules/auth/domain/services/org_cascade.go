package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// OrgCascade lets the auth module trigger organization deletion without importing
// the organization module. Implemented in cmd/wiring/auth/org_cascade_adapter.go.
// Mirrors the existing OrganizationDirectory adapter pattern (design D9).
type OrgCascade interface {
	// CascadeSolelyOwnedOrgs deletes every org the user solely owns, in full.
	// Returns ErrBillingActive (surfaced as 409) if any org is blocked by an
	// active paid subscription — in that case NO further orgs are processed and
	// the user is NOT deleted (caller must resolve billing first).
	CascadeSolelyOwnedOrgs(ctx context.Context, userID uuid.UUID) error
}

// ErrBillingActive is re-exported in auth's domain so the auth handler can map it
// to 409 without importing the organization module. The adapter in
// cmd/wiring/auth/org_cascade_adapter.go translates the org-domain sentinel into
// this one (DAO-16, design D9).
var ErrBillingActive = errors.New("organization has an active paid subscription")
