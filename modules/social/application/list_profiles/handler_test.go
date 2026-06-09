package listprofiles_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories/mocks"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"

	listprofiles "github.com/jcsoftdev/pulzifi-back/modules/social/application/list_profiles"
)

func TestListProfilesHandler(t *testing.T) {
	workspaceID := uuid.New()
	now := time.Now().UTC()

	makeProfile := func(handle string) *entities.SocialProfile {
		return &entities.SocialProfile{
			ID:                   uuid.New(),
			WorkspaceID:          workspaceID,
			Platform:             valueobjects.PlatformInstagram,
			Handle:               handle,
			IsActive:             true,
			CheckIntervalMinutes: 720,
			NextCheckAt:          &now,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
	}

	t.Run("returns profiles scoped to workspace", func(t *testing.T) {
		profiles := []*entities.SocialProfile{makeProfile("nasa"), makeProfile("spacex")}
		repo := &repomocks.MockProfileRepository{
			ListByWorkspaceResult: profiles,
		}
		h := listprofiles.NewHandler(repo)

		resp, err := h.Handle(context.Background(), workspaceID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp) != 2 {
			t.Errorf("expected 2 profiles, got %d", len(resp))
		}
		if repo.ListByWorkspaceCalls != 1 {
			t.Errorf("expected 1 ListByWorkspace call, got %d", repo.ListByWorkspaceCalls)
		}
	})

	t.Run("empty workspace returns empty slice", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{
			ListByWorkspaceResult: nil,
		}
		h := listprofiles.NewHandler(repo)

		resp, err := h.Handle(context.Background(), workspaceID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Error("expected non-nil empty slice, got nil")
		}
		if len(resp) != 0 {
			t.Errorf("expected 0 profiles, got %d", len(resp))
		}
	})

	t.Run("repo error propagated", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{
			ListByWorkspaceErr: context.DeadlineExceeded,
		}
		h := listprofiles.NewHandler(repo)

		_, err := h.Handle(context.Background(), workspaceID)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
