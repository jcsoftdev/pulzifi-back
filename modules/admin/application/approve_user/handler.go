package approveuser

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
	adminerrors "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	authentities "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/entities"
	authrepos "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/repositories"
	orgentities "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/entities"
	orgrepos "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/repositories"
	orgservices "github.com/jcsoftdev/pulzifi-back/modules/organization/domain/services"
	sharedDB "github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// Handler handles user approval
type Handler struct {
	db         *sql.DB
	regReqRepo repositories.RegistrationRequestRepository
	userRepo   authrepos.UserRepository
	orgRepo    orgrepos.OrganizationRepository
	orgService *orgservices.OrganizationService
}

// NewHandler creates a new handler instance
func NewHandler(
	db *sql.DB,
	regReqRepo repositories.RegistrationRequestRepository,
	userRepo authrepos.UserRepository,
	orgRepo orgrepos.OrganizationRepository,
	orgService *orgservices.OrganizationService,
) *Handler {
	return &Handler{
		db:         db,
		regReqRepo: regReqRepo,
		userRepo:   userRepo,
		orgRepo:    orgRepo,
		orgService: orgService,
	}
}

// Handle executes the approve user use case
func (h *Handler) Handle(ctx context.Context, requestID uuid.UUID, reviewerID uuid.UUID) error {
	// Get the registration request
	regReq, err := h.regReqRepo.GetByID(ctx, requestID)
	if err != nil {
		logger.Error("Failed to get registration request", zap.Error(err))
		return err
	}
	if regReq == nil {
		return adminerrors.ErrRegistrationRequestNotFound
	}

	if regReq.Status != entities.RegistrationStatusPending {
		return adminerrors.ErrAlreadyReviewed
	}

	// Execute everything in a transaction
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		logger.Error("Failed to begin transaction", zap.Error(err))
		return err
	}
	defer tx.Rollback()

	// 1. Update user status to approved
	_, err = tx.ExecContext(ctx,
		`UPDATE public.users SET status = $1, updated_at = NOW() WHERE id = $2`,
		authentities.UserStatusApproved, regReq.UserID,
	)
	if err != nil {
		logger.Error("Failed to update user status", zap.Error(err))
		return err
	}

	// 2. Update registration request status to approved
	_, err = tx.ExecContext(ctx,
		`UPDATE public.registration_requests SET status = $1, reviewed_by = $2, reviewed_at = NOW(), updated_at = NOW() WHERE id = $3`,
		entities.RegistrationStatusApproved, reviewerID, requestID,
	)
	if err != nil {
		logger.Error("Failed to update registration request status", zap.Error(err))
		return err
	}

	// 3. Generate schema name and create organization
	schemaName := h.orgService.GenerateSchemaName(regReq.OrganizationSubdomain)
	org := orgentities.NewOrganization(regReq.OrganizationName, regReq.OrganizationSubdomain, schemaName, regReq.UserID)

	_, err = tx.ExecContext(ctx,
		`INSERT INTO public.organizations (id, name, subdomain, schema_name, owner_user_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		org.ID, org.Name, org.Subdomain, org.SchemaName, org.OwnerUserID, org.CreatedAt, org.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to create organization", zap.Error(err))
		return fmt.Errorf("failed to create organization: %w", err)
	}

	// 4. Insert organization member (role: owner)
	_, err = tx.ExecContext(ctx,
		`INSERT INTO public.organization_members (id, organization_id, user_id, role, joined_at) VALUES ($1, $2, $3, $4, NOW())`,
		uuid.New(), org.ID, regReq.UserID, "owner",
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
		regReq.UserID, adminRoleID,
	)
	if err != nil {
		logger.Error("Failed to assign ADMIN role", zap.Error(err))
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// 6. Assign default (starter) plan to the organization
	_, err = tx.ExecContext(ctx,
		`INSERT INTO public.organization_plans (id, organization_id, plan_id, status, started_at, created_at, updated_at)
		 SELECT gen_random_uuid(), $1, id, 'active', NOW(), NOW(), NOW()
		 FROM public.plans WHERE name = 'starter' LIMIT 1`,
		org.ID,
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
	if err := sharedDB.ProvisionTenantSchema(h.db, schemaName); err != nil {
		logger.Error("Failed to provision tenant schema after approval — manual migration may be needed",
			zap.Error(err),
			zap.String("schema", schemaName),
		)
	}

	logger.Info("User approved successfully",
		zap.String("user_id", regReq.UserID.String()),
		zap.String("org_subdomain", regReq.OrganizationSubdomain),
		zap.String("schema_name", schemaName),
	)

	return nil
}

// GetRegistrationRequest retrieves a registration request by ID.
func (h *Handler) GetRegistrationRequest(ctx context.Context, requestID uuid.UUID) (*entities.RegistrationRequest, error) {
	return h.regReqRepo.GetByID(ctx, requestID)
}
