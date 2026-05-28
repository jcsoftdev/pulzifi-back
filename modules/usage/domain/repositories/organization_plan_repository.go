package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/entities"
)

// Tx is an abstraction over *sql.Tx so the domain layer does not import database/sql.
type Tx interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (interface{}, error)
	Rollback() error
	Commit() error
}

// OrgWithPlan is a projection for the list-organizations-with-plans query.
type OrgWithPlan struct {
	ID                uuid.UUID
	Name              string
	Subdomain         string
	SchemaName        string
	PlanCode          string
	PlanName          string
	ChecksAllowed     int
	StoragePeriodDays int
}

// OrgPlanInfo contains plan details needed for billing operations.
type OrgPlanInfo struct {
	SchemaName        string
	ChecksAllowed     int
	StartedAt         interface{} // time.Time from DB
	StoragePeriodDays int
	// AIInsightsAllowed mirrors plans.ai_insights_allowed_monthly. NULL in the
	// plan (Enterprise = unlimited) is resolved to the INT max sentinel
	// (2147483647) by the repository so callers never see a raw NULL.
	AIInsightsAllowed int
}

// OrganizationPlanRepository manages organization_plans in the public schema.
// WithTx returns a transaction-scoped copy for atomic operations.
type OrganizationPlanRepository interface {
	// ListOrgsWithPlans returns all organizations with their active plan details.
	ListOrgsWithPlans(ctx context.Context) ([]*OrgWithPlan, error)

	// DeactivateActive marks the current active plan for an org as inactive.
	DeactivateActive(ctx context.Context, orgID uuid.UUID) error

	// Insert creates a new organization_plan row.
	Insert(ctx context.Context, plan *entities.OrganizationPlan) error

	// GetSchemaName returns the schema_name for an organization.
	GetSchemaName(ctx context.Context, orgID uuid.UUID) (string, error)

	// GetActivePlanForOrg returns active plan details for billing (gift_month, get_quotas).
	GetActivePlanForOrg(ctx context.Context, orgID uuid.UUID) (*OrgPlanInfo, error)

	// GetActivePlanForTenant returns plan details for a tenant schema (used in get_quotas).
	GetActivePlanForTenant(ctx context.Context, tenant string) (*OrgPlanInfo, error)

	// WithTx returns a copy of the repository bound to the given transaction.
	WithTx(tx interface{}) OrganizationPlanRepository
}
