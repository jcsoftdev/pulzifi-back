package deleteprofile_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	domainerrors "github.com/jcsoftdev/pulzifi-back/modules/social/domain/errors"
	repomocks "github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories/mocks"

	deleteprofile "github.com/jcsoftdev/pulzifi-back/modules/social/application/delete_profile"
)

func TestDeleteProfileHandler(t *testing.T) {
	profileID := uuid.New()

	t.Run("happy path — profile deleted", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{}
		h := deleteprofile.NewHandler(repo)

		err := h.Handle(context.Background(), profileID)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.DeleteCalls != 1 {
			t.Errorf("expected 1 Delete call, got %d", repo.DeleteCalls)
		}
	})

	t.Run("404 when profile not found", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{
			DeleteErr: domainerrors.ErrProfileNotFound,
		}
		h := deleteprofile.NewHandler(repo)

		err := h.Handle(context.Background(), profileID)

		if !errors.Is(err, domainerrors.ErrProfileNotFound) {
			t.Errorf("expected ErrProfileNotFound, got %v", err)
		}
	})

	t.Run("repo error propagated", func(t *testing.T) {
		repo := &repomocks.MockProfileRepository{
			DeleteErr: errors.New("db error"),
		}
		h := deleteprofile.NewHandler(repo)

		err := h.Handle(context.Background(), profileID)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
