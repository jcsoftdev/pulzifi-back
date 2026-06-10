package getprofile_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/social/domain/errors"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories/mocks"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"

	getprofile "github.com/jcsoftdev/pulzifi-back/modules/social/application/get_profile"
)

func TestGetProfileHandler(t *testing.T) {
	profileID := uuid.New()
	workspaceID := uuid.New()
	now := time.Now().UTC()

	makeProfile := func() *entities.SocialProfile {
		return &entities.SocialProfile{
			ID:                   profileID,
			WorkspaceID:          workspaceID,
			Platform:             valueobjects.PlatformInstagram,
			Handle:               "nasa",
			IsActive:             true,
			CheckIntervalMinutes: 720,
			NextCheckAt:          &now,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
	}

	makeSnapshot := func() *entities.SocialSnapshot {
		return &entities.SocialSnapshot{
			ID:             uuid.New(),
			ProfileID:      profileID,
			CapturedAt:     now,
			Status:         entities.SnapshotStatusSuccess,
			FollowersCount: 5000,
			PostsCount:     42,
		}
	}

	t.Run("happy path — returns profile and latest snapshot", func(t *testing.T) {
		snapshot := makeSnapshot()
		profileRepo := &repomocks.MockProfileRepository{
			GetByIDResult: makeProfile(),
		}
		snapshotRepo := &repomocks.MockSnapshotRepository{
			GetLatestByProfileResult: snapshot,
		}
		h := getprofile.NewHandler(profileRepo, snapshotRepo)

		resp, err := h.Handle(context.Background(), profileID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if resp.Profile.ID != profileID {
			t.Errorf("profile ID mismatch: want %v, got %v", profileID, resp.Profile.ID)
		}
		if resp.LatestSnapshot == nil {
			t.Error("expected latest snapshot to be set")
		}
		if resp.LatestSnapshot.FollowersCount != 5000 {
			t.Errorf("followers count: want 5000, got %d", resp.LatestSnapshot.FollowersCount)
		}
	})

	t.Run("returns profile with nil snapshot when no snapshot exists yet", func(t *testing.T) {
		profileRepo := &repomocks.MockProfileRepository{
			GetByIDResult: makeProfile(),
		}
		snapshotRepo := &repomocks.MockSnapshotRepository{
			GetLatestByProfileResult: nil, // no snapshot yet
		}
		h := getprofile.NewHandler(profileRepo, snapshotRepo)

		resp, err := h.Handle(context.Background(), profileID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.LatestSnapshot != nil {
			t.Error("expected nil snapshot for profile with no checks")
		}
	})

	t.Run("404 on not found", func(t *testing.T) {
		profileRepo := &repomocks.MockProfileRepository{
			GetByIDErr: domainerrors.ErrProfileNotFound,
		}
		snapshotRepo := &repomocks.MockSnapshotRepository{}
		h := getprofile.NewHandler(profileRepo, snapshotRepo)

		_, err := h.Handle(context.Background(), profileID)

		if !errors.Is(err, domainerrors.ErrProfileNotFound) {
			t.Errorf("expected ErrProfileNotFound, got %v", err)
		}
	})
}
