package adminwiring

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	adminservices "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/services"
	authentities "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	orgentities "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/entities"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	sharedDB "github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// approvalProvisioner implements adminservices.ApprovalProvisioner.
// It encapsulates the full approval transaction (user status update,
// registration request update, org create/reuse, membership, role assignment,
// plan assignment) plus the post-commit schema provisioning.
type approvalProvisioner struct {
	db         *sql.DB
	orgService *orgservices.OrganizationService
}

// NewApprovalProvisioner creates an ApprovalProvisioner backed by *sql.DB and
// the organization domain service (for schema name generation).
func NewApprovalProvisioner(db *sql.DB, orgService *orgservices.OrganizationService) adminservices.ApprovalProvisioner {
	return &approvalProvisioner{db: db, orgService: orgService}
}

func (p *approvalProvisioner) Provision(ctx context.Context, input adminservices.ApprovalInput) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Failed to begin approval transaction", zap.Error(err))
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// 1. Update user status to approved
	_, err = tx.ExecContext(ctx,
		`UPDATE public.users SET status = $1, updated_at = NOW() WHERE id = $2`,
		authentities.UserStatusApproved, input.UserID,
	)
	if err != nil {
		logger.Error("Failed to update user status", zap.Error(err))
		return err
	}

	// 2. Update registration request status to approved
	_, err = tx.ExecContext(ctx,
		`UPDATE public.registration_requests SET status = $1, reviewed_by = $2, reviewed_at = NOW(), updated_at = NOW() WHERE id = $3`,
		"approved", input.ReviewerID, input.RequestID,
	)
	if err != nil {
		logger.Error("Failed to update registration request status", zap.Error(err))
		return err
	}

	// 3. Create organization — or reuse if already approved for this subdomain
	schemaName := p.orgService.GenerateSchemaName(input.OrganizationSubdomain)
	var orgID uuid.UUID
	orgAlreadyExisted := false

	err = tx.QueryRowContext(ctx,
		`SELECT id, schema_name FROM public.organizations WHERE subdomain = $1`,
		input.OrganizationSubdomain,
	).Scan(&orgID, &schemaName)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		// Org doesn't exist yet — create it
		org := orgentities.NewOrganization(input.OrganizationName, input.OrganizationSubdomain, schemaName, input.UserID)
		orgID = org.ID
		_, err = tx.ExecContext(ctx,
			`INSERT INTO public.organizations (id, name, subdomain, schema_name, owner_user_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			org.ID, org.Name, org.Subdomain, org.SchemaName, org.OwnerUserID, org.CreatedAt, org.UpdatedAt,
		)
		if err != nil {
			logger.Error("Failed to create organization", zap.Error(err))
			return fmt.Errorf("failed to create organization: %w", err)
		}
	case err != nil:
		logger.Error("Failed to check for existing organization", zap.Error(err))
		return fmt.Errorf("failed to check organization: %w", err)
	default:
		orgAlreadyExisted = true
		logger.Info("Organization already exists, adding user as member",
			zap.String("subdomain", input.OrganizationSubdomain),
			zap.String("user_id", input.UserID.String()),
		)
	}

	// 4. Insert organization member (owner if new org, MEMBER if joining an existing one)
	memberRole := "owner"
	if orgAlreadyExisted {
		memberRole = "MEMBER"
	}
	_, err = tx.ExecContext(ctx,
		`INSERT INTO public.organization_members (id, organization_id, user_id, role, joined_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 ON CONFLICT (organization_id, user_id) DO NOTHING`,
		uuid.New(), orgID, input.UserID, memberRole,
	)
	if err != nil {
		logger.Error("Failed to create organization member", zap.Error(err))
		return fmt.Errorf("failed to create organization member: %w", err)
	}

	// 5. Assign ADMIN role to user
	var adminRoleID uuid.UUID
	err = tx.QueryRowContext(ctx, `SELECT id FROM public.roles WHERE name = 'ADMIN' LIMIT 1`).Scan(&adminRoleID)
	if err != nil {
		logger.Error("Failed to find ADMIN role", zap.Error(err))
		return fmt.Errorf("failed to find ADMIN role: %w", err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO public.user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		input.UserID, adminRoleID,
	)
	if err != nil {
		logger.Error("Failed to assign ADMIN role", zap.Error(err))
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// 6. Assign default (starter) plan to the organization (only if newly created)
	_, err = tx.ExecContext(ctx,
		`INSERT INTO public.organization_plans (id, organization_id, plan_id, status, started_at, created_at, updated_at)
		 SELECT gen_random_uuid(), $1, id, 'active', NOW(), NOW(), NOW()
		 FROM public.plans WHERE name = 'starter' LIMIT 1
		 ON CONFLICT DO NOTHING`,
		orgID,
	)
	if err != nil {
		logger.Error("Failed to assign default plan", zap.Error(err))
		return fmt.Errorf("failed to assign default plan: %w", err)
	}

	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit approval transaction", zap.Error(err))
		return err
	}

	// Provision tenant schema (DDL — must run outside the transaction)
	// Skip if the org already existed — schema was already provisioned during the first approval
	if !orgAlreadyExisted {
		if err := sharedDB.ProvisionTenantSchema(p.db, schemaName); err != nil {
			logger.Error("Failed to provision tenant schema after approval — manual migration may be needed",
				zap.Error(err),
				zap.String("schema", schemaName),
			)
		}
	}

	logger.Info("User approved successfully",
		zap.String("user_id", input.UserID.String()),
		zap.String("org_subdomain", input.OrganizationSubdomain),
		zap.String("schema_name", schemaName),
	)

	return nil
}
