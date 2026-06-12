package deleteorganization

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

const (
	stripeCancelTimeout  = 15 * time.Second
	storageSweepTimeout  = 30 * time.Second
	schemaDropTimeout    = 30 * time.Second
)

// Handler orchestrates the full organization deletion cascade.
// It is intentionally synchronous: the caller gets a final result before returning.
type Handler struct {
	repo    services.OrgDeletionRepo
	billing services.BillingCanceller
	storage services.StorageSweeper
	schema  services.SchemaDropper
}

// NewHandler constructs a Handler with all required ports.
func NewHandler(
	repo services.OrgDeletionRepo,
	billing services.BillingCanceller,
	storage services.StorageSweeper,
	schema services.SchemaDropper,
) *Handler {
	return &Handler{
		repo:    repo,
		billing: billing,
		storage: storage,
		schema:  schema,
	}
}

// Handle runs the full cascade for a single org.
//
// Returns ErrOrgNotFound (→ 404), ErrBillingActive (→ 409), or nil (→ 204).
// Every failure path marks the audit row before returning — the caller never
// sees an unrecorded failure state.
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	log := logger.Logger.With(
		zap.String("org_id", req.OrgID.String()),
		zap.String("actor_id", req.ActorID.String()),
		zap.String("actor_type", req.ActorType),
	)

	// ── Step 1: Lookup ────────────────────────────────────────────────────────
	log.Info("delete_organization: starting", zap.String("step", "lookup"))
	org, err := h.repo.LookupForDeletion(ctx, req.OrgID)
	if err != nil {
		if errors.Is(err, services.ErrOrgNotFound) {
			return nil, services.ErrOrgNotFound
		}
		return nil, fmt.Errorf("delete_organization: lookup: %w", err)
	}

	log = log.With(zap.String("schema_name", org.SchemaName))

	// ── Step 2: TX-A — soft-delete + open audit ───────────────────────────────
	log.Info("delete_organization: soft-delete + open audit", zap.String("step", "soft_delete"))
	auditID, err := h.repo.SoftDeleteAndOpenAudit(ctx, services.AuditOpenInput{
		OrgID:     req.OrgID,
		Schema:    org.SchemaName,
		ActorType: req.ActorType,
		ActorID:   req.ActorID,
	})
	if err != nil {
		return nil, fmt.Errorf("delete_organization: soft_delete_and_open_audit: %w", err)
	}

	// ── Step 3: Stripe cancellation ───────────────────────────────────────────
	log.Info("delete_organization: billing cancellation", zap.String("step", "billing"))
	stripeCtx, stripeCancel := context.WithTimeout(ctx, stripeCancelTimeout)
	defer stripeCancel()

	if err := h.billing.CancelForOrg(stripeCtx, req.OrgID); err != nil {
		if errors.Is(err, services.ErrBillingActive) {
			log.Error("delete_organization: billing active, aborting",
				zap.String("step", "billing"),
				zap.Error(err),
			)
			_ = h.repo.MarkAudit(ctx, auditID, "failed", "billing", err.Error())
			return nil, services.ErrBillingActive
		}
		// Non-billing errors from the canceller are also fatal
		_ = h.repo.MarkAudit(ctx, auditID, "failed", "billing", err.Error())
		return nil, fmt.Errorf("delete_organization: billing cancel: %w", err)
	}

	// ── Step 4: Storage sweep (best-effort, never aborts) ─────────────────────
	log.Info("delete_organization: storage sweep", zap.String("step", "storage_sweep"))
	sweepCtx, sweepCancel := context.WithTimeout(ctx, storageSweepTimeout)
	defer sweepCancel()

	if err := h.storage.SweepTenant(sweepCtx, req.OrgID, org.SchemaName); err != nil {
		log.Warn("delete_organization: storage sweep failed (continuing)",
			zap.String("step", "storage_sweep"),
			zap.String("org_id", req.OrgID.String()),
			zap.String("schema_name", org.SchemaName),
			zap.Error(err),
		)
		// Do NOT abort — storage sweep is best-effort.
	}

	// ── Step 5: TX-B — cleanup + hard-delete ─────────────────────────────────
	log.Info("delete_organization: hard-delete", zap.String("step", "hard_delete"))
	deletedUserIDs, err := h.repo.CleanupAndHardDelete(ctx, req.OrgID, org.SchemaName)
	if err != nil {
		log.Error("delete_organization: hard-delete failed",
			zap.String("step", "hard_delete"),
			zap.Error(err),
		)
		_ = h.repo.MarkAudit(ctx, auditID, "failed", "hard_delete", err.Error())
		return nil, fmt.Errorf("delete_organization: cleanup_and_hard_delete: %w", err)
	}

	// ── Step 6: DROP SCHEMA ───────────────────────────────────────────────────
	log.Info("delete_organization: drop schema", zap.String("step", "drop_schema"))
	dropCtx, dropCancel := context.WithTimeout(ctx, schemaDropTimeout)
	defer dropCancel()

	if err := h.schema.DropTenantSchema(dropCtx, org.SchemaName); err != nil {
		log.Error("delete_organization: drop schema failed",
			zap.String("step", "drop_schema"),
			zap.String("schema_name", org.SchemaName),
			zap.Error(err),
		)
		_ = h.repo.MarkAudit(ctx, auditID, "failed", "drop_schema", err.Error())
		return nil, fmt.Errorf("delete_organization: drop_schema: %w", err)
	}

	// ── Step 7: Mark audit completed ─────────────────────────────────────────
	if err := h.repo.MarkAudit(ctx, auditID, "completed", "", ""); err != nil {
		// Non-fatal: data is gone, only audit is not closed. Log and continue.
		log.Error("delete_organization: failed to mark audit completed",
			zap.String("audit_id", auditID.String()),
			zap.Error(err),
		)
	}

	log.Info("delete_organization: completed",
		zap.Strings("deleted_user_ids", uuidSliceToStrings(deletedUserIDs)),
	)

	return &Response{
		OrgID:          req.OrgID,
		Schema:         org.SchemaName,
		DeletedUserIDs: deletedUserIDs,
	}, nil
}

func uuidSliceToStrings(ids []uuid.UUID) []string {
	s := make([]string, len(ids))
	for i, id := range ids {
		s[i] = id.String()
	}
	return s
}
