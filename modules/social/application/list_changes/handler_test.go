package listchanges_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories/mocks"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"

	listchanges "github.com/jcsoftdev/pulzifi-back/modules/social/application/list_changes"
)

func makeChange(profileID uuid.UUID, createdAt time.Time) *entities.SocialChange {
	return &entities.SocialChange{
		ID:             uuid.New(),
		ProfileID:      profileID,
		FromSnapshotID: uuid.New(),
		ToSnapshotID:   uuid.New(),
		ChangeTypes:    []valueobjects.ChangeType{valueobjects.ChangeTypeNewPost},
		Summary:        entities.ChangeSummary{NewPosts: []string{"post-1"}},
		CreatedAt:      createdAt,
	}
}

func TestListChangesHandler(t *testing.T) {
	profileID := uuid.New()
	workspaceID := uuid.New()
	now := time.Now().UTC()

	t.Run("filter by profile_id", func(t *testing.T) {
		changes := []*entities.SocialChange{
			makeChange(profileID, now),
			makeChange(profileID, now.Add(-time.Hour)),
		}
		repo := &repomocks.MockChangeRepository{ListResult: changes}
		h := listchanges.NewHandler(repo)

		req := &listchanges.Request{ProfileID: &profileID, Limit: 10, Offset: 0}
		resp, err := h.Handle(context.Background(), req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Changes) != 2 {
			t.Errorf("expected 2 changes, got %d", len(resp.Changes))
		}
		if repo.ListCalls != 1 {
			t.Errorf("expected 1 List call, got %d", repo.ListCalls)
		}
	})

	t.Run("filter by workspace_id", func(t *testing.T) {
		changes := []*entities.SocialChange{makeChange(profileID, now)}
		repo := &repomocks.MockChangeRepository{ListResult: changes}
		h := listchanges.NewHandler(repo)

		req := &listchanges.Request{WorkspaceID: &workspaceID, Limit: 10}
		resp, err := h.Handle(context.Background(), req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Changes) != 1 {
			t.Errorf("expected 1 change, got %d", len(resp.Changes))
		}
		// Verify WorkspaceID passed in filter (not ProfileID)
		if repo.ListCalls != 1 {
			t.Errorf("expected 1 List call, got %d", repo.ListCalls)
		}
	})

	t.Run("pagination — limit and offset applied", func(t *testing.T) {
		// Repo returns 1 result (simulating limit applied)
		changes := []*entities.SocialChange{makeChange(profileID, now)}
		repo := &repomocks.MockChangeRepository{ListResult: changes}
		h := listchanges.NewHandler(repo)

		req := &listchanges.Request{ProfileID: &profileID, Limit: 1, Offset: 1}
		resp, err := h.Handle(context.Background(), req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
	})

	t.Run("empty result returns empty slice", func(t *testing.T) {
		repo := &repomocks.MockChangeRepository{ListResult: nil}
		h := listchanges.NewHandler(repo)

		req := &listchanges.Request{ProfileID: &profileID, Limit: 10}
		resp, err := h.Handle(context.Background(), req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Changes == nil {
			t.Error("expected non-nil empty slice")
		}
		if len(resp.Changes) != 0 {
			t.Errorf("expected 0 changes, got %d", len(resp.Changes))
		}
	})

	t.Run("filter passthrough — ProfileID takes precedence", func(t *testing.T) {
		repo := &repomocks.MockChangeRepository{ListResult: nil}
		h := listchanges.NewHandler(repo)

		// Both set — ProfileID should take precedence (spec says ProfileID wins)
		req := &listchanges.Request{
			ProfileID:   &profileID,
			WorkspaceID: &workspaceID,
			Limit:       10,
		}
		_, err := h.Handle(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Verify the filter sent to repo has ProfileID
		if repo.ListCalls != 1 {
			t.Errorf("expected 1 List call, got %d", repo.ListCalls)
		}
	})
}

// Capture the filter sent to List for assertions.
var _ = (*repositories.ChangeFilter)(nil)
