package listplans_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/application/list_plans"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories/mocks"
)

func TestListPlansHandler_Handle(t *testing.T) {
	planA := &entities.Plan{ID: uuid.New(), Code: "starter", Name: "Starter", ChecksAllowedMonthly: 100, IsActive: true, StoragePeriodDays: 7}
	planB := &entities.Plan{ID: uuid.New(), Code: "pro", Name: "Pro", ChecksAllowedMonthly: 1000, IsActive: true, StoragePeriodDays: 30}

	tests := []struct {
		name    string
		setup   func(*mocks.MockPlanRepository)
		wantLen int
		wantErr bool
	}{
		{
			name: "returns all active plans",
			setup: func(m *mocks.MockPlanRepository) {
				m.ListActiveResult = []*entities.Plan{planA, planB}
			},
			wantLen: 2,
		},
		{
			name: "returns empty list when no plans",
			setup: func(m *mocks.MockPlanRepository) {
				m.ListActiveResult = nil
			},
			wantLen: 0,
		},
		{
			name: "propagates repo error",
			setup: func(m *mocks.MockPlanRepository) {
				m.ListActiveErr = errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockPlanRepository{}
			tt.setup(repo)

			h := listplans.NewHandler(repo)
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
			if len(resp.Plans) != tt.wantLen {
				t.Errorf("plans count: want %d, got %d", tt.wantLen, len(resp.Plans))
			}
		})
	}
}
