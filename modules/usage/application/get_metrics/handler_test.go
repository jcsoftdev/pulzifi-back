package getmetrics_test

import (
	"context"
	"errors"
	"testing"
	"time"

	getmetrics "github.com/jcsoftdev/pulzifi-back/modules/usage/application/get_metrics"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories/mocks"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/services"
)

func TestGetMetricsHandler_Handle(t *testing.T) {
	billing := services.New()

	tests := []struct {
		name           string
		setup          func(*mocks.MockUsageTrackingRepository, *mocks.MockOrganizationPlanRepository)
		wantChecksUsed int
		wantPages      int
		wantErr        bool
	}{
		{
			name: "returns metrics with active period",
			setup: func(u *mocks.MockUsageTrackingRepository, o *mocks.MockOrganizationPlanRepository) {
				u.TotalChecks = 50
				u.SuccessChecks = 40
				u.FailedChecks = 10
				u.TotalPages = 5
				u.TotalWorkspaces = 2
				u.TotalAlerts = 3
				u.FindCurrentResult = &entities.UsageTracking{ChecksUsed: 25, ChecksAllowed: 100}
			},
			wantChecksUsed: 25,
			wantPages:      5,
		},
		{
			name: "returns zero metrics when no active period",
			setup: func(u *mocks.MockUsageTrackingRepository, o *mocks.MockOrganizationPlanRepository) {
				u.FindCurrentResult = nil
				u.FindCurrentErr = errors.New("no period")
			},
			wantChecksUsed: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usageRepo := &mocks.MockUsageTrackingRepository{}
			orgRepo := &mocks.MockOrganizationPlanRepository{}
			tt.setup(usageRepo, orgRepo)

			h := getmetrics.NewHandler(usageRepo, orgRepo, billing)
			resp, err := h.Handle(context.Background(), &getmetrics.Request{Tenant: "acme"})

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
			if resp.Pages != tt.wantPages {
				t.Errorf("pages: want %d, got %d", tt.wantPages, resp.Pages)
			}
		})
	}
}

// ensure billing service param is used (compile-time check for struct usage)
var _ = time.Now
