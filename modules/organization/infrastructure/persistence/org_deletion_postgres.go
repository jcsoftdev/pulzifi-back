package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// orgDeletionPostgres implements services.OrgDeletionRepo against the public schema.
// All operations use fully-qualified table names; no SET search_path required.
type orgDeletionPostgres struct {
	db *sql.DB
}

// NewOrgDeletionPostgres constructs a production-ready OrgDeletionRepo.
func NewOrgDeletionPostgres(db *sql.DB) services.OrgDeletionRepo {
	return &orgDeletionPostgres{db: db}
}

// LookupForDeletion returns org identity by ID, including soft-deleted orgs
// (needed for idempotent re-run). Returns ErrOrgNotFound when the row is absent
// (already hard-deleted or never existed).
func (r *orgDeletionPostgres) LookupForDeletion(ctx context.Context, orgID uuid.UUID) (services.OrgForDeletion, error) {
	var (
		id         uuid.UUID
		schemaName string
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT id, schema_name
		  FROM public.organizations
		 WHERE id = $1
	`, orgID).Scan(&id, &schemaName)
	if errors.Is(err, sql.ErrNoRows) {
		return services.OrgForDeletion{}, services.ErrOrgNotFound
	}
	if err != nil {
		return services.OrgForDeletion{}, fmt.Errorf("org_deletion_postgres: lookup: %w", err)
	}
	return services.OrgForDeletion{ID: id, SchemaName: schemaName}, nil
}

// SoftDeleteAndOpenAudit executes TX-A: soft-deletes the org and inserts the
// audit row in a single transaction. Idempotent: if the org is already
// soft-deleted it still inserts (or reuses) an audit row.
func (r *orgDeletionPostgres) SoftDeleteAndOpenAudit(ctx context.Context, in services.AuditOpenInput) (uuid.UUID, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, fmt.Errorf("org_deletion_postgres: begin tx-a: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Soft-delete — no-op if already soft-deleted (WHERE deleted_at IS NULL guard).
	_, err = tx.ExecContext(ctx, `
		UPDATE public.organizations
		   SET deleted_at = NOW()
		 WHERE id = $1
		   AND deleted_at IS NULL
	`, in.OrgID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("org_deletion_postgres: soft_delete: %w", err)
	}

	// Check for an existing pending or failed audit row (idempotent re-run).
	// A prior run that ended with status='failed' must be resumed, not duplicated.
	// We reset it to 'pending' and clear failure metadata so the cascade re-runs
	// from this point forward.
	var existingAuditID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM public.organization_deletions
		 WHERE organization_id = $1
		   AND status IN ('pending', 'failed')
		 ORDER BY requested_at DESC
		 LIMIT 1
	`, in.OrgID).Scan(&existingAuditID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, fmt.Errorf("org_deletion_postgres: check existing audit: %w", err)
	}
	if err == nil {
		// Reuse existing audit row: reset status to pending and clear failure fields.
		_, err = tx.ExecContext(ctx, `
			UPDATE public.organization_deletions
			   SET status        = 'pending',
			       failure_step  = NULL,
			       error_message = NULL,
			       completed_at  = NULL
			 WHERE id = $1
		`, existingAuditID)
		if err != nil {
			return uuid.Nil, fmt.Errorf("org_deletion_postgres: reset existing audit: %w", err)
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return uuid.Nil, fmt.Errorf("org_deletion_postgres: commit tx-a (reuse): %w", commitErr)
		}
		return existingAuditID, nil
	}
	// No existing pending row — insert fresh. UUID generated in Go (OQ-4).
	auditID := uuid.New()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO public.organization_deletions
		       (id, organization_id, schema_name, actor_type, actor_id, status, requested_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', NOW())
	`, auditID, in.OrgID, in.Schema, in.ActorType, in.ActorID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("org_deletion_postgres: insert audit: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return uuid.Nil, fmt.Errorf("org_deletion_postgres: commit tx-a: %w", err)
	}
	return auditID, nil
}

// MarkAudit updates the audit row status and optional failure metadata.
func (r *orgDeletionPostgres) MarkAudit(ctx context.Context, auditID uuid.UUID, status, failureStep, errMsg string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE public.organization_deletions
		   SET status        = $2,
		       failure_step  = NULLIF($3, ''),
		       error_message = NULLIF($4, ''),
		       completed_at  = CASE WHEN $2 IN ('completed', 'failed') THEN NOW() ELSE NULL END
		 WHERE id = $1
	`, auditID, status, failureStep, errMsg)
	if err != nil {
		return fmt.Errorf("org_deletion_postgres: mark_audit: %w", err)
	}
	return nil
}

// CleanupAndHardDelete executes TX-B:
//  1. Delete outbox_events + outbox_consumed rows for the tenant.
//  2. Hard-delete the org row (FK cascades members/plans/integrations/quotas).
//  3. For each former member with zero remaining memberships, hard-delete the user.
//
// Returns the list of user IDs that were hard-deleted.
func (r *orgDeletionPostgres) CleanupAndHardDelete(ctx context.Context, orgID uuid.UUID, schema string) ([]uuid.UUID, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("org_deletion_postgres: begin tx-b: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1a. Collect member user IDs before the org row is deleted.
	rows, err := tx.QueryContext(ctx, `
		SELECT user_id FROM public.organization_members
		 WHERE organization_id = $1
		   AND deleted_at IS NULL
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("org_deletion_postgres: list members: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var memberIDs []uuid.UUID
	for rows.Next() {
		var uid uuid.UUID
		if scanErr := rows.Scan(&uid); scanErr != nil {
			return nil, fmt.Errorf("org_deletion_postgres: scan member: %w", scanErr)
		}
		memberIDs = append(memberIDs, uid)
	}
	if closeErr := rows.Close(); closeErr != nil {
		return nil, fmt.Errorf("org_deletion_postgres: close member rows: %w", closeErr)
	}

	// 1b. Delete outbox_consumed rows that reference events for this tenant.
	_, err = tx.ExecContext(ctx, `
		DELETE FROM public.outbox_consumed
		 WHERE event_id IN (
		   SELECT id FROM public.outbox_events WHERE tenant = $1
		 )
	`, schema)
	if err != nil {
		return nil, fmt.Errorf("org_deletion_postgres: delete outbox_consumed: %w", err)
	}

	// 1c. Delete outbox_events for this tenant.
	_, err = tx.ExecContext(ctx, `
		DELETE FROM public.outbox_events WHERE tenant = $1
	`, schema)
	if err != nil {
		return nil, fmt.Errorf("org_deletion_postgres: delete outbox_events: %w", err)
	}

	// 2. Hard-delete org (FK cascades members, plans, integrations, quotas).
	_, err = tx.ExecContext(ctx, `
		DELETE FROM public.organizations WHERE id = $1
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("org_deletion_postgres: hard_delete_org: %w", err)
	}

	// 3. For each former member, delete user if no remaining memberships.
	// Non-fatal pruning errors use a loop-local pruneErr to avoid polluting the
	// named return err that the defer rollback closure reads. Clearing err = nil
	// inside the loop would cause the deferred rollback to be skipped if the last
	// iteration errored and err was left as nil before Commit.
	var deletedUserIDs []uuid.UUID
	for _, userID := range memberIDs {
		var remaining int
		pruneErr := tx.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM public.organization_members
			 WHERE user_id    = $1
			   AND deleted_at IS NULL
		`, userID).Scan(&remaining)
		if pruneErr != nil {
			logger.Warn("org_deletion_postgres: count memberships failed, skipping user prune",
				zap.String("user_id", userID.String()),
				zap.Error(pruneErr),
			)
			continue
		}
		if remaining == 0 {
			_, pruneErr = tx.ExecContext(ctx, `
				DELETE FROM public.users WHERE id = $1
			`, userID)
			if pruneErr != nil {
				logger.Warn("org_deletion_postgres: user hard-delete failed, continuing",
					zap.String("user_id", userID.String()),
					zap.Error(pruneErr),
				)
				continue
			}
			deletedUserIDs = append(deletedUserIDs, userID)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("org_deletion_postgres: commit tx-b: %w", err)
	}
	return deletedUserIDs, nil
}

// FindSolelyOwnedOrgs returns orgs where userID is the sole active owner.
// Mirrors the isSoleOwner logic from the subscriber.
func (r *orgDeletionPostgres) FindSolelyOwnedOrgs(ctx context.Context, userID uuid.UUID) ([]services.OrgForDeletion, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT o.id, o.schema_name
		  FROM public.organizations o
		 WHERE o.deleted_at IS NULL
		   AND EXISTS (
		     SELECT 1 FROM public.organization_members m
		      WHERE m.organization_id = o.id
		        AND m.user_id         = $1
		        AND m.role            = 'OWNER'
		        AND m.deleted_at      IS NULL
		   )
		   AND NOT EXISTS (
		     SELECT 1 FROM public.organization_members m2
		      WHERE m2.organization_id = o.id
		        AND m2.user_id         <> $1
		        AND m2.role            = 'OWNER'
		        AND m2.deleted_at      IS NULL
		   )
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("org_deletion_postgres: find_solely_owned: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []services.OrgForDeletion
	for rows.Next() {
		var org services.OrgForDeletion
		if scanErr := rows.Scan(&org.ID, &org.SchemaName); scanErr != nil {
			return nil, fmt.Errorf("org_deletion_postgres: scan solely_owned: %w", scanErr)
		}
		result = append(result, org)
	}
	return result, rows.Err()
}
