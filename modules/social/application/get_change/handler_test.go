package getchange_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories/mocks"
	valueobjects "github.com/jcsoftdev/pulzifi-back/modules/social/domain/value_objects"

	getchange "github.com/jcsoftdev/pulzifi-back/modules/social/application/get_change"
)

func TestGetChangeHandler(t *testing.T) {
	changeID := uuid.New()
	profileID := uuid.New()
	fromID := uuid.New()
	toID := uuid.New()
	now := time.Now().UTC()

	makeSnapshot := func(id uuid.UUID, data *entities.ProfileData) *entities.SocialSnapshot {
		return &entities.SocialSnapshot{
			ID:             id,
			ProfileID:      profileID,
			CapturedAt:     now,
			Status:         entities.SnapshotStatusSuccess,
			FollowersCount: 1000,
			PostsCount:     10,
			Data:           data,
		}
	}

	prevData := &entities.ProfileData{Bio: "Old bio", FollowersCount: 900}
	nextData := &entities.ProfileData{Bio: "New bio", FollowersCount: 1000}

	makeChange := func() *entities.SocialChange {
		return &entities.SocialChange{
			ID:             changeID,
			ProfileID:      profileID,
			FromSnapshotID: fromID,
			ToSnapshotID:   toID,
			ChangeTypes:    []valueobjects.ChangeType{valueobjects.ChangeTypeBioChanged},
			Summary:        entities.ChangeSummary{Bio: &entities.TextDiff{From: "Old bio", To: "New bio"}},
			CreatedAt:      now,
		}
	}

	t.Run("happy path — returns change with before/after ProfileData", func(t *testing.T) {
		changeRepo := &repomocks.MockChangeRepository{GetByIDResult: makeChange()}
		snapshotRepo := &repomocks.MockSnapshotRepository{}
		// Use Fn overrides to return correct snapshot per ID
		snapshotRepo.GetByIDFn = func(ctx context.Context, id uuid.UUID) (*entities.SocialSnapshot, error) {
			if id == fromID {
				return makeSnapshot(fromID, prevData), nil
			}
			if id == toID {
				return makeSnapshot(toID, nextData), nil
			}
			return nil, nil
		}

		h := getchange.NewHandler(changeRepo, snapshotRepo)

		resp, err := h.Handle(context.Background(), changeID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if resp.Change.ID != changeID {
			t.Errorf("change ID mismatch: want %v, got %v", changeID, resp.Change.ID)
		}
		if resp.BeforeData == nil {
			t.Error("expected non-nil BeforeData")
		}
		if resp.AfterData == nil {
			t.Error("expected non-nil AfterData")
		}
		if resp.BeforeData.Bio != "Old bio" {
			t.Errorf("before bio: want 'Old bio', got %q", resp.BeforeData.Bio)
		}
		if resp.AfterData.Bio != "New bio" {
			t.Errorf("after bio: want 'New bio', got %q", resp.AfterData.Bio)
		}
	})

	t.Run("404 when change not found", func(t *testing.T) {
		changeRepo := &repomocks.MockChangeRepository{GetByIDResult: nil}
		snapshotRepo := &repomocks.MockSnapshotRepository{}
		h := getchange.NewHandler(changeRepo, snapshotRepo)

		resp, err := h.Handle(context.Background(), changeID)

		if err == nil && resp != nil {
			t.Fatal("expected error or nil response for missing change")
		}
	})
}
