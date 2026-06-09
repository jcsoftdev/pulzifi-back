package createprofile_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/social/domain/errors"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories/mocks"
	svcmocks "github.com/jcsoftdev/pulzifi-back/modules/social/domain/services/mocks"

	createprofile "github.com/jcsoftdev/pulzifi-back/modules/social/application/create_profile"
)

func TestCreateProfileHandler(t *testing.T) {
	workspaceID := uuid.New()

	makeHandler := func(
		profileRepo *repomocks.MockProfileRepository,
		planLimits *svcmocks.MockPlanLimits,
	) *createprofile.Handler {
		return createprofile.NewHandler(profileRepo, planLimits)
	}

	t.Run("happy path — profile created", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{
			CountActiveByWorkspaceResult: 0,
		}
		plan := &svcmocks.MockPlanLimits{
			GetProfilesLimitResult: 10, // up to 10 profiles
			GetChecksPerDayResult:  120,
		}
		h := makeHandler(repo, plan)

		resp, err := h.Handle(context.Background(), &createprofile.Request{
			WorkspaceID:          workspaceID,
			Platform:             "instagram",
			Handle:               "nasa",
			CheckIntervalMinutes: 720,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if resp.ID == uuid.Nil {
			t.Error("response ID should not be nil UUID")
		}
		if resp.Handle != "nasa" {
			t.Errorf("handle: want 'nasa', got %q", resp.Handle)
		}
		if resp.Platform != "instagram" {
			t.Errorf("platform: want 'instagram', got %q", resp.Platform)
		}
		if !resp.IsActive {
			t.Error("new profile should be active")
		}
		if resp.NextCheckAt == nil {
			t.Error("next_check_at should be set on creation")
		}
		if repo.CreateCalls != 1 {
			t.Errorf("expected 1 Create call, got %d", repo.CreateCalls)
		}
	})

	t.Run("invalid platform rejected (400)", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{}
		plan := &svcmocks.MockPlanLimits{}
		h := makeHandler(repo, plan)

		_, err := h.Handle(context.Background(), &createprofile.Request{
			WorkspaceID:          workspaceID,
			Platform:             "twitter",
			Handle:               "nasa",
			CheckIntervalMinutes: 720,
		})

		if err == nil {
			t.Fatal("expected validation error for invalid platform")
		}
		if repo.CreateCalls != 0 {
			t.Error("Create should not be called for invalid platform")
		}
	})

	t.Run("empty handle rejected (400)", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{}
		plan := &svcmocks.MockPlanLimits{}
		h := makeHandler(repo, plan)

		_, err := h.Handle(context.Background(), &createprofile.Request{
			WorkspaceID:          workspaceID,
			Platform:             "instagram",
			Handle:               "",
			CheckIntervalMinutes: 720,
		})

		if err == nil {
			t.Fatal("expected error for empty handle")
		}
	})

	t.Run("invalid interval preset rejected (400)", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{}
		plan := &svcmocks.MockPlanLimits{}
		h := makeHandler(repo, plan)

		_, err := h.Handle(context.Background(), &createprofile.Request{
			WorkspaceID:          workspaceID,
			Platform:             "instagram",
			Handle:               "nasa",
			CheckIntervalMinutes: 90, // not a valid preset
		})

		if err == nil {
			t.Fatal("expected error for invalid interval preset")
		}
		if repo.CreateCalls != 0 {
			t.Error("Create should not be called for invalid interval")
		}
	})

	t.Run("plan profile limit reached (422)", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{
			CountActiveByWorkspaceResult: 1, // already at limit
		}
		plan := &svcmocks.MockPlanLimits{
			GetProfilesLimitResult: 1, // max 1 profile
			GetChecksPerDayResult:  120,
		}
		h := makeHandler(repo, plan)

		_, err := h.Handle(context.Background(), &createprofile.Request{
			WorkspaceID:          workspaceID,
			Platform:             "instagram",
			Handle:               "nasa",
			CheckIntervalMinutes: 720,
		})

		if !errors.Is(err, domainerrors.ErrProfileLimitReached) {
			t.Errorf("expected ErrProfileLimitReached, got %v", err)
		}
		if repo.CreateCalls != 0 {
			t.Error("Create should not be called when plan limit reached")
		}
	})

	t.Run("duplicate handle rejected (409)", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{
			CountActiveByWorkspaceResult: 0,
			CreateErr:                    domainerrors.ErrProfileAlreadyExists,
		}
		plan := &svcmocks.MockPlanLimits{
			GetProfilesLimitResult: 10,
			GetChecksPerDayResult:  120,
		}
		h := makeHandler(repo, plan)

		_, err := h.Handle(context.Background(), &createprofile.Request{
			WorkspaceID:          workspaceID,
			Platform:             "instagram",
			Handle:               "nasa",
			CheckIntervalMinutes: 720,
		})

		if !errors.Is(err, domainerrors.ErrProfileAlreadyExists) {
			t.Errorf("expected ErrProfileAlreadyExists, got %v", err)
		}
	})

	t.Run("interval exceeds plan checks per day (422)", func(t *testing.T) {
		// Plan: 2 checks/day, 1 active profile at 720min interval
		// → current demand = 1 * (1440/720) = 2 checks/day
		// Adding another at 720min = 2 * (1440/720) = 4 → exceeds 2
		repo := &repomocks.MockProfileRepository{
			CountActiveByWorkspaceResult: 1, // 1 existing active profile
		}
		plan := &svcmocks.MockPlanLimits{
			GetProfilesLimitResult: 10,
			GetChecksPerDayResult:  2, // only 2 checks/day on free tier
		}
		h := makeHandler(repo, plan)

		_, err := h.Handle(context.Background(), &createprofile.Request{
			WorkspaceID:          workspaceID,
			Platform:             "instagram",
			Handle:               "nasa",
			CheckIntervalMinutes: 720, // 2 checks/day per profile
		})

		if err == nil {
			t.Fatal("expected error: interval would exceed plan capacity")
		}
		if repo.CreateCalls != 0 {
			t.Error("Create should not be called when interval exceeds plan")
		}
	})

	t.Run("unlimited plan (0) — no profile limit enforced", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{
			CountActiveByWorkspaceResult: 100,
		}
		plan := &svcmocks.MockPlanLimits{
			GetProfilesLimitResult: 0, // 0 = unlimited
			GetChecksPerDayResult:  0, // 0 = unlimited
		}
		h := makeHandler(repo, plan)

		resp, err := h.Handle(context.Background(), &createprofile.Request{
			WorkspaceID:          workspaceID,
			Platform:             "instagram",
			Handle:               "nasa",
			CheckIntervalMinutes: 120,
		})

		if err != nil {
			t.Fatalf("unexpected error for unlimited plan: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
	})

	t.Run("feature disabled for plan (-1 profile limit)", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{}
		plan := &svcmocks.MockPlanLimits{
			GetProfilesLimitResult: -1, // -1 = feature disabled
		}
		h := makeHandler(repo, plan)

		_, err := h.Handle(context.Background(), &createprofile.Request{
			WorkspaceID:          workspaceID,
			Platform:             "instagram",
			Handle:               "nasa",
			CheckIntervalMinutes: 720,
		})

		if !errors.Is(err, domainerrors.ErrProfileLimitReached) {
			t.Errorf("expected ErrProfileLimitReached for disabled plan, got %v", err)
		}
	})
}
