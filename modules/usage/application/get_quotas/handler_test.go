package getquotas_test

import (
	"context"
	"errors"
	"testing"
	"time"

	getquotas "github.com/jcsoftdev/pulzifi-back/modules/usage/application/get_quotas"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories/mocks"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/services"
)

func TestGetQuotasHandler_Handle(t *testing.T) {
	billing := services.New()
	refillTime := time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name              string
		setup             func(*mocks.MockUsageTrackingRepository, *mocks.MockOrganizationPlanRepository)
		wantChecksUsed    int
		wantAllowed       int
		wantMaxPages      int
		wantMaxWorkspaces int
		wantErr           bool
	}{
		{
			name: "returns quota from existing period with plan limits",
			setup: func(u *mocks.MockUsageTrackingRepository, o *mocks.MockOrganizationPlanRepository) {
				u.FindCurrentResult = &entities.UsageTracking{
					ChecksUsed:        300,
					ChecksAllowed:     1000,
					NextRefillAt:      &refillTime,
					StoragePeriodDays: 7,
				}
				o.GetActivePlanTenResult = &repositories.OrgPlanInfo{
					ChecksAllowed:     1000,
					StoragePeriodDays: 7,
					MaxPages:          45,
					MaxWorkspaces:     10,
					StartedAt:         time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				}
			},
			wantChecksUsed:    300,
			wantAllowed:       1000,
			wantMaxPages:      45,
			wantMaxWorkspaces: 10,
		},
		{
			name: "creates period from plan when none exists",
			setup: func(u *mocks.MockUsageTrackingRepository, o *mocks.MockOrganizationPlanRepository) {
				u.FindCurrentResult = nil
				o.GetActivePlanTenResult = &repositories.OrgPlanInfo{
					ChecksAllowed:     500,
					StoragePeriodDays: 14,
					MaxPages:          22,
					MaxWorkspaces:     1,
					StartedAt:         time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				}
			},
			wantChecksUsed:    0,
			wantAllowed:       500,
			wantMaxPages:      22,
			wantMaxWorkspaces: 1,
		},
		{
			name: "unlimited plan returns sentinel for max_pages and max_workspaces",
			setup: func(u *mocks.MockUsageTrackingRepository, o *mocks.MockOrganizationPlanRepository) {
				u.FindCurrentResult = nil
				o.GetActivePlanTenResult = &repositories.OrgPlanInfo{
					ChecksAllowed:     500,
					StoragePeriodDays: 30,
					MaxPages:          2147483647, // NULL resolved to sentinel by repo
					MaxWorkspaces:     2147483647,
					StartedAt:         time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
				}
			},
			wantChecksUsed:    0,
			wantAllowed:       500,
			wantMaxPages:      2147483647,
			wantMaxWorkspaces: 2147483647,
		},
		{
			name: "plan fetch error with existing period degrades gracefully to sentinel",
			setup: func(u *mocks.MockUsageTrackingRepository, o *mocks.MockOrganizationPlanRepository) {
				u.FindCurrentResult = &entities.UsageTracking{
					ChecksUsed:        100,
					ChecksAllowed:     500,
					NextRefillAt:      &refillTime,
					StoragePeriodDays: 7,
				}
				o.GetActivePlanTenErr = errors.New("db error")
			},
			wantChecksUsed:    100,
			wantAllowed:       500,
			wantMaxPages:      2147483647,
			wantMaxWorkspaces: 2147483647,
		},
		{
			name: "error when find current fails",
			setup: func(u *mocks.MockUsageTrackingRepository, o *mocks.MockOrganizationPlanRepository) {
				u.FindCurrentErr = errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "error when no active plan and no period",
			setup: func(u *mocks.MockUsageTrackingRepository, o *mocks.MockOrganizationPlanRepository) {
				u.FindCurrentResult = nil
				o.GetActivePlanTenResult = nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usageRepo := &mocks.MockUsageTrackingRepository{}
			orgRepo := &mocks.MockOrganizationPlanRepository{}
			tt.setup(usageRepo, orgRepo)

			h := getquotas.NewHandler(usageRepo, orgRepo, billing)
			resp, err := h.Handle(context.Background(), &getquotas.Request{Tenant: "acme"})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.ChecksUsed != tt.wantChecksUsed {
				t.Errorf("checks_used: want %d, got %d", tt.wantChecksUsed, resp.ChecksUsed)
			}
			if resp.ChecksAllowed != tt.wantAllowed {
				t.Errorf("checks_allowed: want %d, got %d", tt.wantAllowed, resp.ChecksAllowed)
			}
			if resp.MaxPages != tt.wantMaxPages {
				t.Errorf("max_pages: want %d, got %d", tt.wantMaxPages, resp.MaxPages)
			}
			if resp.MaxWorkspaces != tt.wantMaxWorkspaces {
				t.Errorf("max_workspaces: want %d, got %d", tt.wantMaxWorkspaces, resp.MaxWorkspaces)
			}
		})
	}
}
