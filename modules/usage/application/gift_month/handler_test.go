package giftmonth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	giftmonth "github.com/jcsoftdev/pulzifi-back/modules/usage/application/gift_month"
	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/repositories/mocks"
	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/services"
)

func TestGiftMonthHandler_Handle(t *testing.T) {
	billing := services.New()
	orgID := uuid.New()
	startedAt := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	latestEnd := time.Date(2025, 3, 14, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		setup   func(*mocks.MockOrganizationPlanRepository, *mocks.MockUsageTrackingRepository)
		wantErr bool
	}{
		{
			name: "successfully gifts a month",
			setup: func(o *mocks.MockOrganizationPlanRepository, u *mocks.MockUsageTrackingRepository) {
				o.GetActivePlanResult = &repositories.OrgPlanInfo{
					SchemaName:        "acme",
					ChecksAllowed:     1000,
					StoragePeriodDays: 7,
					StartedAt:         startedAt,
				}
				u.LatestPeriodEndResult = latestEnd
			},
		},
		{
			name: "error when no active plan",
			setup: func(o *mocks.MockOrganizationPlanRepository, u *mocks.MockUsageTrackingRepository) {
				o.GetActivePlanResult = nil
			},
			wantErr: true,
		},
		{
			name: "error when latest period lookup fails",
			setup: func(o *mocks.MockOrganizationPlanRepository, u *mocks.MockUsageTrackingRepository) {
				o.GetActivePlanResult = &repositories.OrgPlanInfo{
					SchemaName:    "acme",
					ChecksAllowed: 1000,
					StartedAt:     startedAt,
				}
				u.LatestPeriodEndErr = errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orgRepo := &mocks.MockOrganizationPlanRepository{}
			usageRepo := &mocks.MockUsageTrackingRepository{}
			tt.setup(orgRepo, usageRepo)

			h := giftmonth.NewHandler(orgRepo, usageRepo, billing)
			resp, err := h.Handle(context.Background(), &giftmonth.Request{OrgID: orgID})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.OrganizationID != orgID {
				t.Errorf("org_id: want %v, got %v", orgID, resp.OrganizationID)
			}
			if resp.GiftedPeriod.ChecksAllowed != 1000 {
				t.Errorf("checks_allowed: want 1000, got %d", resp.GiftedPeriod.ChecksAllowed)
			}
		})
	}
}
