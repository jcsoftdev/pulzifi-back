package assignplan_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	assignplan "github.com/jcsoftdev/pulzifi-back/modules/usage/application/assign_plan"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories/mocks"
)

func TestAssignPlanHandler_Handle(t *testing.T) {
	orgID := uuid.New()

	tests := []struct {
		name    string
		setup   func(*mocks.MockPlanRepository, *mocks.MockOrganizationPlanRepository, *mocks.MockUsageTrackingRepository)
		wantErr bool
	}{
		{
			name: "plan not found returns error",
			setup: func(p *mocks.MockPlanRepository, o *mocks.MockOrganizationPlanRepository, u *mocks.MockUsageTrackingRepository) {
				p.GetByCodeResult = nil // nil == not found
			},
			wantErr: true,
		},
		{
			name: "get plan error propagates",
			setup: func(p *mocks.MockPlanRepository, o *mocks.MockOrganizationPlanRepository, u *mocks.MockUsageTrackingRepository) {
				p.GetByCodeErr = errors.New("db error")
			},
			wantErr: true,
		},
	}

	// Note: the happy path test (commit path) requires a real *sql.DB with BeginTx support.
	// We test it at the integration level. The unit tests here focus on pre-tx error paths
	// where we can use a nil DB (BeginTx is never called).

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			planRepo := &mocks.MockPlanRepository{}
			orgRepo := &mocks.MockOrganizationPlanRepository{}
			usageRepo := &mocks.MockUsageTrackingRepository{}
			tt.setup(planRepo, orgRepo, usageRepo)

			txb := &mocks.MockTxBeginner{}
			h := assignplan.NewHandler(txb, planRepo, orgRepo, func(string) repositories.UsageTrackingRepository { return usageRepo })
			_, err := h.Handle(context.Background(), &assignplan.Request{OrgID: orgID, PlanCode: "pro"})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// TestAssignPlanHandler_PlanNotFound checks the plan not found path explicitly.
func TestAssignPlanHandler_PlanNotFound(t *testing.T) {
	planRepo := &mocks.MockPlanRepository{GetByCodeResult: nil}
	orgRepo := &mocks.MockOrganizationPlanRepository{}
	usageRepo := &mocks.MockUsageTrackingRepository{}

	txb := &mocks.MockTxBeginner{}
	h := assignplan.NewHandler(txb, planRepo, orgRepo, func(string) repositories.UsageTrackingRepository { return usageRepo })
	_, err := h.Handle(context.Background(), &assignplan.Request{OrgID: uuid.New(), PlanCode: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for plan not found, got nil")
	}
}
