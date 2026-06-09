package getquotastatus_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	svcmocks "github.com/jcsoftdev/pulzifi-back/modules/social/domain/services/mocks"

	getquotastatus "github.com/jcsoftdev/pulzifi-back/modules/social/application/get_quota_status"
)

func TestGetQuotaStatusHandler(t *testing.T) {
	workspaceID := uuid.New()

	t.Run("limited plan — returns used/limit/resets_at", func(t *testing.T) {
		plan := &svcmocks.MockPlanLimits{
			GetChecksPerDayResult: 120,
			GetChecksUsedResult:   12,
		}
		h := getquotastatus.NewHandler(plan)

		resp, err := h.Handle(context.Background(), workspaceID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if resp.ChecksUsed != 12 {
			t.Errorf("checks_used: want 12, got %d", resp.ChecksUsed)
		}
		if resp.ChecksLimit == nil {
			t.Fatal("checks_limit should be non-nil for limited plan")
		}
		if *resp.ChecksLimit != 120 {
			t.Errorf("checks_limit: want 120, got %d", *resp.ChecksLimit)
		}
		if resp.ResetsAt.IsZero() {
			t.Error("resets_at should not be zero")
		}
		// resets_at should be the next UTC midnight
		nextMidnight := nextUTCMidnight()
		diff := resp.ResetsAt.Sub(nextMidnight)
		if diff < -time.Minute || diff > time.Minute {
			t.Errorf("resets_at is not approximately next UTC midnight: got %v", resp.ResetsAt)
		}
	})

	t.Run("unlimited plan (0) — limit is null, used still returned", func(t *testing.T) {
		plan := &svcmocks.MockPlanLimits{
			GetChecksPerDayResult: 0, // 0 = unlimited
			GetChecksUsedResult:   5,
		}
		h := getquotastatus.NewHandler(plan)

		resp, err := h.Handle(context.Background(), workspaceID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.ChecksLimit != nil {
			t.Errorf("checks_limit should be nil for unlimited plan, got %v", resp.ChecksLimit)
		}
		if resp.ChecksUsed != 5 {
			t.Errorf("checks_used: want 5, got %d", resp.ChecksUsed)
		}
	})

	t.Run("read-only — does not modify quota", func(t *testing.T) {
		plan := &svcmocks.MockPlanLimits{
			GetChecksPerDayResult: 120,
			GetChecksUsedResult:   0,
		}
		h := getquotastatus.NewHandler(plan)

		_, err := h.Handle(context.Background(), workspaceID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// GetChecksUsedToday called once; GetChecksPerDay called once
		if plan.GetChecksUsedCalls != 1 {
			t.Errorf("expected 1 GetChecksUsedToday call, got %d", plan.GetChecksUsedCalls)
		}
		if plan.GetChecksPerDayCalls != 1 {
			t.Errorf("expected 1 GetChecksPerDay call, got %d", plan.GetChecksPerDayCalls)
		}
	})
}

func nextUTCMidnight() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
}
