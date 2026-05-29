// Package authwiring contains cross-module adapters that bridge the auth module
// to other modules (admin, organization, email) without creating module-to-module
// import cycles. This file implements TrialProvisioner — the new self-serve
// onboarding path that replaces the manual admin approval gate.
package authwiring

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	authentities "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	orgentities "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/entities"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	sharedDB "github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// schemaProvisioner is the seam that lets tests skip real DDL on the database.
type schemaProvisioner func(db *sql.DB, schemaName string) error

// trialProvisioner implements authservices.TrialProvisioner.
//
// It runs a single transaction that:
//  1. Inserts the organization (status implicit via row creation).
//  2. Adds the user as the owner member.
//  3. Assigns the ADMIN role.
//  4. Creates the active organization_plans row for the trial plan with
//     trial_ends_at = now() + trial_days.
//
// After the transaction commits, the tenant schema is provisioned (DDL must
// run outside the transaction). usage_tracking is seeded by the tenant
// migration 000004_seed_usage_tracking_from_plan which reads the active plan
// row — so the plan row MUST exist before ProvisionTenantSchema runs.
type trialProvisioner struct {
	db         *sql.DB
	orgService *orgservices.OrganizationService
	provision  schemaProvisioner
	nowFn      func() time.Time
}

// NewTrialProvisioner constructs the production TrialProvisioner that issues
// real DDL via shared/database.ProvisionTenantSchema.
func NewTrialProvisioner(db *sql.DB, orgService *orgservices.OrganizationService) authservices.TrialProvisioner {
	return &trialProvisioner{
		db:         db,
		orgService: orgService,
		provision:  sharedDB.ProvisionTenantSchema,
		nowFn:      time.Now,
	}
}

// Provision executes the self-serve onboarding transaction.
func (p *trialProvisioner) Provision(ctx context.Context, in authservices.TrialProvisionInput) (*authservices.TrialProvisionOutput, error) {
	if in.TrialDays <= 0 {
		in.TrialDays = 14
	}

	schemaName := p.orgService.GenerateSchemaName(in.OrganizationSubdomain)
	org := orgentities.NewOrganization(in.OrganizationName, in.OrganizationSubdomain, schemaName, in.UserID)

	trialEndsAt := p.nowFn().Add(time.Duration(in.TrialDays) * 24 * time.Hour)

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Failed to begin trial provisioning transaction", zap.Error(err))
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1. Create organization
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO public.organizations (id, name, subdomain, schema_name, owner_user_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		org.ID, org.Name, org.Subdomain, org.SchemaName, org.OwnerUserID, org.CreatedAt, org.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("trial provisioner: insert organization: %w", err)
	}

	// 2. Add user as owner member
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO public.organization_members (id, organization_id, user_id, role, joined_at)
		 VALUES ($1, $2, $3, 'owner', NOW())`,
		uuid.New(), org.ID, in.UserID,
	); err != nil {
		return nil, fmt.Errorf("trial provisioner: insert member: %w", err)
	}

	// 3. Assign ADMIN role to the new owner
	var adminRoleID uuid.UUID
	if err = tx.QueryRowContext(ctx,
		`SELECT id FROM public.roles WHERE name = 'ADMIN' LIMIT 1`,
	).Scan(&adminRoleID); err != nil {
		return nil, fmt.Errorf("trial provisioner: find ADMIN role: %w", err)
	}
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO public.user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		in.UserID, adminRoleID,
	); err != nil {
		return nil, fmt.Errorf("trial provisioner: assign ADMIN role: %w", err)
	}

	// 4. Insert active organization_plans row for the trial plan
	if _, err = tx.ExecContext(ctx,
		`INSERT INTO public.organization_plans
		    (id, organization_id, plan_id, status, started_at, trial_ends_at, created_at, updated_at)
		 SELECT gen_random_uuid(), $1, p.id, 'active', NOW(), $2, NOW(), NOW()
		   FROM public.plans p
		  WHERE p.code = 'trial' AND p.is_active = TRUE
		  LIMIT 1`,
		org.ID, trialEndsAt,
	); err != nil {
		return nil, fmt.Errorf("trial provisioner: insert trial plan: %w", err)
	}

	if err = tx.Commit(); err != nil {
		logger.Error("Failed to commit trial provisioning transaction", zap.Error(err))
		return nil, err
	}

	// Provision tenant schema (DDL — must run outside the transaction).
	if err = p.provision(p.db, schemaName); err != nil {
		logger.Error("Failed to provision tenant schema after trial signup — manual migration may be needed",
			zap.Error(err),
			zap.String("schema", schemaName),
		)
		// Do not roll back the parent commit — the organization row is real;
		// schema provisioning can be retried out-of-band.
	}

	// Ensure the User entity constants stay referenced (avoid unused-import
	// drift if downstream changes touch the user status).
	_ = authentities.UserStatusApproved

	return &authservices.TrialProvisionOutput{
		OrganizationID:        org.ID,
		OrganizationSubdomain: org.Subdomain,
		SchemaName:            schemaName,
		TrialEndsAt:           trialEndsAt,
	}, nil
}
