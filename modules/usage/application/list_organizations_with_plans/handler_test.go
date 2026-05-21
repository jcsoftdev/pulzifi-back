package listorganizationswithplans_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	listorgs "github.com/jcsoftdev/pulzifi-back/modules/usage/application/list_organizations_with_plans"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories/mocks"
)

func TestListOrgsWithPlansHandler_Handle(t *testing.T) {
	orgA := &repositories.OrgWithPlan{ID: uuid.New(), Name: "Acme", Subdomain: "acme", SchemaName: "acme_schema", PlanCode: "pro", PlanName: "Pro", ChecksAllowed: 1000, StoragePeriodDays: 30}

	tests := []struct {
		name    string
		setup   func(*mocks.MockOrganizationPlanRepository)
		wantLen int
		wantErr bool
	}{
		{
			name: "returns list of orgs",
			setup: func(m *mocks.MockOrganizationPlanRepository) {
				m.ListOrgsResult = []*repositories.OrgWithPlan{orgA}
			},
			wantLen: 1,
		},
		{
			name: "returns empty list",
			setup: func(m *mocks.MockOrganizationPlanRepository) {
				m.ListOrgsResult = nil
			},
			wantLen: 0,
		},
		{
			name: "propagates repo error",
			setup: func(m *mocks.MockOrganizationPlanRepository) {
				m.ListOrgsErr = errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockOrganizationPlanRepository{}
			tt.setup(repo)

			h := listorgs.NewHandler(repo)
			resp, err := h.Handle(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(resp.Organizations) != tt.wantLen {
				t.Errorf("orgs count: want %d, got %d", tt.wantLen, len(resp.Organizations))
			}
			if tt.wantLen > 0 && resp.Organizations[0].PlanCode != orgA.PlanCode {
				t.Errorf("plan_code: want %q, got %q", orgA.PlanCode, resp.Organizations[0].PlanCode)
			}
		})
	}
}
