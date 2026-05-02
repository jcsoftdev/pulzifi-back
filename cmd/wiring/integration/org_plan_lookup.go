package intwiring

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

// OrgPlanLookup resolves the active plan code for an organization.
// Used by Twilio provider to determine tier (free / paid / enterprise).
type OrgPlanLookup struct {
	db *sql.DB
}

func NewOrgPlanLookup(db *sql.DB) *OrgPlanLookup {
	return &OrgPlanLookup{db: db}
}

// PlanCode returns the org's currently-active plan code.
// Returns ("", nil) if the org has no active plan (treated as free tier upstream).
func (l *OrgPlanLookup) PlanCode(ctx context.Context, orgID uuid.UUID) (string, error) {
	var code string
	err := l.db.QueryRowContext(ctx, `
		SELECT p.code
		  FROM public.organization_plans op
		  JOIN public.plans p ON p.id = op.plan_id
		 WHERE op.organization_id = $1 AND op.status = 'active' AND op.deleted_at IS NULL
		 ORDER BY op.started_at DESC
		 LIMIT 1
	`, orgID).Scan(&code)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return code, err
}
