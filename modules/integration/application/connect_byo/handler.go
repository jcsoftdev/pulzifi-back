package connectbyo

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

// Validator verifies that pasted credentials work against the provider.
type Validator interface {
	Validate(ctx context.Context, provider string, creds map[string]string) error
}

// FlagReader checks per-org feature flags.
type FlagReader interface {
	IsOn(ctx context.Context, orgID uuid.UUID, key string) (bool, error)
}

// PlanLookup returns the org's active plan code.
type PlanLookup interface {
	PlanCode(ctx context.Context, orgID uuid.UUID) (string, error)
}

type Request struct {
	Provider    string
	OrgID       uuid.UUID
	UserID      uuid.UUID
	Credentials map[string]string
}

type Response struct {
	Integration *entities.Integration
}

type Handler struct {
	repo      repositories.IntegrationRepository
	validator Validator
	flags     FlagReader
	plans     PlanLookup
	paidPlans []string
}

func NewHandler(repo repositories.IntegrationRepository, validator Validator, flags FlagReader, plans PlanLookup, paidPlans []string) *Handler {
	return &Handler{repo: repo, validator: validator, flags: flags, plans: plans, paidPlans: paidPlans}
}

func (h *Handler) Handle(ctx context.Context, req Request) (*Response, error) {
	// Phase 2 only ships Twilio BYO. Other providers may be added in future phases.
	if req.Provider != "twilio" {
		return nil, errors.New("connect_byo: provider not supported")
	}

	// Flag gate: provider must be enabled for this org.
	on, err := h.flags.IsOn(ctx, req.OrgID, "integrations."+req.Provider)
	if err != nil {
		return nil, fmt.Errorf("connect_byo: flag check: %w", err)
	}
	if !on {
		return nil, errors.New("connect_byo: provider not enabled for this organization")
	}

	// Tier gate: BYO is enterprise-only. Paid orgs use platform creds; free orgs are rejected.
	code, err := h.plans.PlanCode(ctx, req.OrgID)
	if err != nil {
		return nil, fmt.Errorf("connect_byo: plan lookup: %w", err)
	}
	for _, p := range h.paidPlans {
		if p == code {
			return nil, errors.New("connect_byo: BYO only available on enterprise plans")
		}
	}
	if code == "" {
		return nil, errors.New("connect_byo: BYO only available on enterprise plans")
	}
	// code is non-empty AND not in paidPlans → enterprise tier (or another non-paid named plan).

	// Validate creds against the live provider API.
	if err := h.validator.Validate(ctx, req.Provider, req.Credentials); err != nil {
		return nil, fmt.Errorf("connect_byo: invalid credentials: %w", err)
	}

	// Persist. Twilio D4 mapping: AccessToken=AuthToken, RefreshToken=AccountSID, ProviderMeta.from_number.
	integ := &entities.Integration{
		ID:           uuid.New(),
		OrgID:        req.OrgID,
		ServiceType:  req.Provider,
		Status:       entities.IntegrationActive,
		AccessToken:  req.Credentials["auth_token"],
		RefreshToken: req.Credentials["account_sid"],
		ProviderMeta: map[string]any{"from_number": req.Credentials["from_number"]},
		CreatedBy:    req.UserID,
	}
	if err := h.repo.Create(ctx, integ); err != nil {
		return nil, fmt.Errorf("connect_byo: persist: %w", err)
	}
	return &Response{Integration: integ}, nil
}
