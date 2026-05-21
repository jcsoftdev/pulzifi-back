package billingwiring

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	authentities "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	billingservices "github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
)

// Compile-time check.
var _ billingservices.TrialConverter = (*TrialConverter)(nil)

// TrialConverter implements billingservices.TrialConverter via raw SQL on the
// public schema. It is wired in cmd/server/modules.go and called by the
// billing webhook handler after PlanAssigner.Assign on checkout completion.
type TrialConverter struct {
	db *sql.DB
}

// NewTrialConverter constructs the production TrialConverter.
func NewTrialConverter(db *sql.DB) *TrialConverter { return &TrialConverter{db: db} }

// Convert marks every active trial organization_plans row for the org as
// converted, and flips any trial_expired members back to approved. Operations
// run in a single transaction so the two mutations are atomic.
func (c *TrialConverter) Convert(ctx context.Context, orgID uuid.UUID) error {
	if orgID == uuid.Nil {
		return fmt.Errorf("trial converter: orgID is uuid.Nil")
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("trial converter: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1. Mark trial organization_plans rows as converted (skips rows already converted).
	if _, err = tx.ExecContext(ctx, `
		UPDATE public.organization_plans
		   SET converted_at = NOW(),
		       updated_at   = NOW()
		 WHERE organization_id = $1
		   AND converted_at IS NULL
		   AND plan_id IN (SELECT id FROM public.plans WHERE code = 'trial')
	`, orgID); err != nil {
		return fmt.Errorf("trial converter: mark converted: %w", err)
	}

	// 2. Flip every member's user.status back to approved if they were trial_expired.
	if _, err = tx.ExecContext(ctx, `
		UPDATE public.users u
		   SET status     = $2,
		       updated_at = NOW()
		  FROM public.organization_members om
		 WHERE om.user_id = u.id
		   AND om.organization_id = $1
		   AND u.status = $3
	`, orgID, authentities.UserStatusApproved, authentities.UserStatusTrialExpired); err != nil {
		return fmt.Errorf("trial converter: restore user statuses: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("trial converter: commit: %w", err)
	}
	return nil
}
