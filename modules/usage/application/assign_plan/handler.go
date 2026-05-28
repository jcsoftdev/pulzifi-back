package assignplan

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

var schemaNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Handler handles the assign_plan use case.
type Handler struct {
	txBeginner   repositories.TxBeginner
	plans        repositories.PlanRepository
	orgPlans     repositories.OrganizationPlanRepository
	usageFactory repositories.UsageTrackingRepositoryFactory
}

// NewHandler creates a new assign plan handler. usageFactory builds the
// tenant-scoped repo for the TARGET org's schema (resolved post-commit).
func NewHandler(
	txBeginner repositories.TxBeginner,
	plans repositories.PlanRepository,
	orgPlans repositories.OrganizationPlanRepository,
	usageFactory repositories.UsageTrackingRepositoryFactory,
) *Handler {
	return &Handler{txBeginner: txBeginner, plans: plans, orgPlans: orgPlans, usageFactory: usageFactory}
}

// Handle executes the multi-step plan assignment atomically.
func (h *Handler) Handle(ctx context.Context, req *Request) (*Response, error) {
	// Look up the target plan
	plan, err := h.plans.GetByCode(ctx, req.PlanCode)
	if err != nil {
		return nil, fmt.Errorf("get plan: %w", err)
	}
	if plan == nil {
		return nil, fmt.Errorf("plan not found")
	}

	tx, err := h.txBeginner.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer h.txBeginner.Rollback(tx)

	txOrgPlans := h.orgPlans.WithTx(tx)

	// Deactivate current plan
	if err := txOrgPlans.DeactivateActive(ctx, req.OrgID); err != nil {
		return nil, fmt.Errorf("deactivate active plan: %w", err)
	}

	// Insert new plan
	var actorID *uuid.UUID
	if req.ActorUserID != "" {
		parsed, err := uuid.Parse(req.ActorUserID)
		if err == nil {
			actorID = &parsed
		}
	}
	if err := txOrgPlans.Insert(ctx, &entities.OrganizationPlan{
		ID:             uuid.New(),
		OrganizationID: req.OrgID,
		PlanID:         plan.ID,
		Status:         "active",
		CreatedBy:      actorID,
	}); err != nil {
		return nil, fmt.Errorf("insert plan: %w", err)
	}

	// Get schema name for post-commit sync
	schemaName, err := txOrgPlans.GetSchemaName(ctx, req.OrgID)
	if err != nil || schemaName == "" {
		return nil, fmt.Errorf("organization not found")
	}
	if !schemaNameRegex.MatchString(schemaName) {
		return nil, fmt.Errorf("invalid organization schema")
	}

	if err := h.txBeginner.Commit(tx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	// Post-commit best-effort: sync the TARGET org's tenant usage_tracking
	// (non-fatal — tenant may not have the table yet).
	if syncErr := h.usageFactory(schemaName).SyncChecksAllowed(ctx, plan.ChecksAllowedMonthly); syncErr != nil {
		logger.Warn("assign_plan: failed to sync tenant usage_tracking (non-fatal)",
			zap.String("schema", schemaName), zap.Error(syncErr))
	}

	return &Response{
		OrganizationID:       req.OrgID,
		PlanCode:             req.PlanCode,
		ChecksAllowedMonthly: plan.ChecksAllowedMonthly,
	}, nil
}
