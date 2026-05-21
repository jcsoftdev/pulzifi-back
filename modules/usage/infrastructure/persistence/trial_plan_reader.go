package persistence

import (
	"context"
	"database/sql"
	"errors"

	trialstatus "github.com/jcsoftdev/pulzifi-back/modules/usage/application/trial_status"
)

// TrialPlanPostgresReader implements trialstatus.OrgPlanReader by joining
// public.organizations to public.organization_plans to public.plans and
// returning the active plan row for the given tenant subdomain.
type TrialPlanPostgresReader struct {
	db *sql.DB
}

// NewTrialPlanPostgresReader constructs the reader.
func NewTrialPlanPostgresReader(db *sql.DB) *TrialPlanPostgresReader {
	return &TrialPlanPostgresReader{db: db}
}

// ActivePlanBySubdomain returns nil when no active plan exists.
func (r *TrialPlanPostgresReader) ActivePlanBySubdomain(ctx context.Context, subdomain string) (*trialstatus.PlanSnapshot, error) {
	snap := &trialstatus.PlanSnapshot{}
	err := r.db.QueryRowContext(ctx, `
		SELECT p.code, op.trial_ends_at, op.converted_at
		  FROM public.organizations o
		  JOIN public.organization_plans op
		    ON op.organization_id = o.id
		   AND op.status = 'active'
		   AND op.deleted_at IS NULL
		  JOIN public.plans p ON p.id = op.plan_id
		 WHERE o.subdomain = $1
		   AND o.deleted_at IS NULL
		 ORDER BY op.started_at DESC
		 LIMIT 1
	`, subdomain).Scan(&snap.PlanCode, &snap.TrialEndsAt, &snap.ConvertedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return snap, nil
}
