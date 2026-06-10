package updateprofile_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/social/domain/errors"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories/mocks"
	svcmocks "github.com/jcsoftdev/pulzifi-back/modules/social/domain/services/mocks"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"

	updateprofile "github.com/jcsoftdev/pulzifi-back/modules/social/application/update_profile"
)

func TestUpdateProfileHandler(t *testing.T) {
	profileID := uuid.New()
	workspaceID := uuid.New()
	now := time.Now().UTC()

	makeProfile := func(isActive bool, interval int) *entities.SocialProfile {
		return &entities.SocialProfile{
			ID:                   profileID,
			WorkspaceID:          workspaceID,
			Platform:             valueobjects.PlatformInstagram,
			Handle:               "nasa",
			IsActive:             isActive,
			CheckIntervalMinutes: interval,
			NextCheckAt:          &now,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
	}

	t.Run("happy path — update interval (valid preset)", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{
			GetByIDResult: makeProfile(true, 720),
		}
		plan := &svcmocks.MockPlanLimits{
			GetChecksPerDayResult:  120,
			GetProfilesLimitResult: 10,
		}
		h := updateprofile.NewHandler(repo, plan)

		resp, err := h.Handle(context.Background(), profileID, &updateprofile.Request{
			CheckIntervalMinutes: intPtr(1440),
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.CheckIntervalMinutes != 1440 {
			t.Errorf("interval not updated: want 1440, got %d", resp.CheckIntervalMinutes)
		}
		if repo.UpdateCalls != 1 {
			t.Errorf("expected 1 Update call, got %d", repo.UpdateCalls)
		}
	})

	t.Run("is_active=false removes from scheduler (NextCheckAt set to nil)", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{
			GetByIDResult: makeProfile(true, 720),
		}
		plan := &svcmocks.MockPlanLimits{GetChecksPerDayResult: 120}
		h := updateprofile.NewHandler(repo, plan)

		resp, err := h.Handle(context.Background(), profileID, &updateprofile.Request{
			IsActive: boolPtr(false),
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.IsActive {
			t.Error("profile should be inactive")
		}
		if resp.NextCheckAt != nil {
			t.Error("NextCheckAt should be nil when deactivated")
		}
	})

	t.Run("is_active=true resets next_check_at to now()", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{
			GetByIDResult: makeProfile(false, 720),
		}
		plan := &svcmocks.MockPlanLimits{GetChecksPerDayResult: 120}
		h := updateprofile.NewHandler(repo, plan)

		resp, err := h.Handle(context.Background(), profileID, &updateprofile.Request{
			IsActive: boolPtr(true),
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !resp.IsActive {
			t.Error("profile should be active")
		}
		if resp.NextCheckAt == nil {
			t.Error("NextCheckAt should be set when re-activated")
		}
	})

	t.Run("invalid interval preset rejected", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{
			GetByIDResult: makeProfile(true, 720),
		}
		plan := &svcmocks.MockPlanLimits{GetChecksPerDayResult: 120}
		h := updateprofile.NewHandler(repo, plan)

		_, err := h.Handle(context.Background(), profileID, &updateprofile.Request{
			CheckIntervalMinutes: intPtr(90), // invalid
		})

		if err == nil {
			t.Fatal("expected error for invalid interval preset")
		}
		if repo.UpdateCalls != 0 {
			t.Error("Update should not be called for invalid interval")
		}
	})

	t.Run("interval that exceeds plan capacity rejected", func(t *testing.T) {
		// Plan: 2 checks/day, 2 active profiles already
		// Changing this profile to 720min → 2 profiles * 2 checks = 4 > 2
		repo := &repomocks.MockProfileRepository{
			GetByIDResult:                makeProfile(true, 1440),
			CountActiveByWorkspaceResult: 2,
		}
		plan := &svcmocks.MockPlanLimits{
			GetChecksPerDayResult:  2,
			GetProfilesLimitResult: 10,
		}
		h := updateprofile.NewHandler(repo, plan)

		_, err := h.Handle(context.Background(), profileID, &updateprofile.Request{
			CheckIntervalMinutes: intPtr(720), // 2/day each × 2 profiles = 4 > 2
		})

		if err == nil {
			t.Fatal("expected error: interval would exceed plan capacity")
		}
	})

	t.Run("profile not found returns ErrProfileNotFound", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{
			GetByIDErr: domainerrors.ErrProfileNotFound,
		}
		plan := &svcmocks.MockPlanLimits{}
		h := updateprofile.NewHandler(repo, plan)

		_, err := h.Handle(context.Background(), profileID, &updateprofile.Request{})

		if !errors.Is(err, domainerrors.ErrProfileNotFound) {
			t.Errorf("expected ErrProfileNotFound, got %v", err)
		}
	})
}

func intPtr(v int) *int    { return &v }
func boolPtr(v bool) *bool { return &v }
