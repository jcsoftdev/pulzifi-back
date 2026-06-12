package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// OrgForDeletion is the minimal org identity the cascade needs.
type OrgForDeletion struct {
	ID         uuid.UUID
	SchemaName string
	OwnerEmail string // for audit/logging only; may be empty
}

// AuditOpenInput carries what the audit row records at open time.
type AuditOpenInput struct {
	OrgID     uuid.UUID
	Schema    string
	ActorType string    // "owner" | "super_admin"
	ActorID   uuid.UUID // user performing the deletion
}

// OrgDeletionRepo covers every public-schema read/write the cascade performs.
// All methods operate on the public schema (no SET search_path).
type OrgDeletionRepo interface {
	// LookupForDeletion returns org identity by ID. Returns ErrOrgNotFound when
	// the row is absent (already hard-deleted or never existed).
	LookupForDeletion(ctx context.Context, orgID uuid.UUID) (OrgForDeletion, error)

	// SoftDeleteAndOpenAudit soft-deletes the org (deleted_at = NOW()) AND inserts
	// the audit row with status='pending' in ONE transaction. Returns the audit row ID.
	// Idempotent on re-run: if already soft-deleted it reuses/updates the open audit row.
	SoftDeleteAndOpenAudit(ctx context.Context, in AuditOpenInput) (auditID uuid.UUID, err error)

	// MarkAudit transitions the audit row to a terminal/intermediate state.
	MarkAudit(ctx context.Context, auditID uuid.UUID, status, failureStep, errMsg string) error

	// CleanupAndHardDelete runs in ONE transaction:
	//   DELETE outbox_events/outbox_consumed WHERE tenant=$schema,
	//   DELETE organizations WHERE id=$orgID (FK cascades members/plans/integrations/quotas),
	//   then for each former member with zero remaining memberships, DELETE the user.
	// Returns the user IDs that were deleted (for logging/audit).
	CleanupAndHardDelete(ctx context.Context, orgID uuid.UUID, schema string) (deletedUserIDs []uuid.UUID, err error)

	// FindSolelyOwnedOrgs returns orgs where userID is the sole active owner —
	// the set the /auth/me owner path must cascade. Mirrors subscriber.isSoleOwner.
	FindSolelyOwnedOrgs(ctx context.Context, userID uuid.UUID) ([]OrgForDeletion, error)
}

// BillingCanceller decides and performs Stripe cancellation for an org.
type BillingCanceller interface {
	// CancelForOrg cancels the org's active subscription immediately when required.
	// Returns ErrBillingActive when an active+paid sub exists and the cancel call
	// failed (caller MUST abort the cascade and return 409). Returns nil when there
	// is nothing to cancel, the sub is trialing/canceled, or cancellation succeeded.
	CancelForOrg(ctx context.Context, orgID uuid.UUID) error
}

// StorageSweeper removes tenant object-storage artifacts. Best-effort: a sweep
// failure is logged but does NOT abort the cascade.
type StorageSweeper interface {
	// SweepTenant deletes all objects under the {schema}/ prefix when private_storage
	// is enabled for the org; no-op (returns nil) otherwise. Never returns an abort error.
	SweepTenant(ctx context.Context, orgID uuid.UUID, schema string) error
}

// SchemaDropper performs the terminal, irreversible schema drop.
type SchemaDropper interface {
	DropTenantSchema(ctx context.Context, schema string) error
}

var (
	// ErrOrgNotFound is returned when the org row does not exist or has already been hard-deleted.
	ErrOrgNotFound = errors.New("organization not found")

	// ErrBillingActive is returned when an active paid subscription exists and cancel failed.
	// Callers must abort the cascade and return HTTP 409.
	ErrBillingActive = errors.New("organization has an active paid subscription")
)
