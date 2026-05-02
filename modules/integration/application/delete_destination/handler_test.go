package deletedestination_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	deletedestination "github.com/jcsoftdev/pulzifi-back/modules/integration/application/delete_destination"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

type stubDestinationRepo struct {
	deleteCalled int
	deletedID    uuid.UUID
}

func (r *stubDestinationRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.deleteCalled++
	r.deletedID = id
	return nil
}
func (r *stubDestinationRepo) Create(_ context.Context, _ *entities.Destination) error { return nil }
func (r *stubDestinationRepo) Update(_ context.Context, _ *entities.Destination) error { return nil }
func (r *stubDestinationRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.Destination, error) {
	return nil, nil
}
func (r *stubDestinationRepo) ListByScope(_ context.Context, _ entities.ScopeType, _ uuid.UUID) ([]*entities.Destination, error) {
	return nil, nil
}
func (r *stubDestinationRepo) ResolveForEvent(_ context.Context, _ string, _ uuid.UUID, _, _ *uuid.UUID) ([]*entities.Destination, error) {
	return nil, nil
}
func (r *stubDestinationRepo) DisableByIntegrationID(_ context.Context, _ uuid.UUID) error {
	return nil
}

func TestDelete_HappyPath(t *testing.T) {
	id := uuid.New()
	repo := &stubDestinationRepo{}
	h := deletedestination.NewHandler(repo)

	_, err := h.Handle(context.Background(), deletedestination.Request{ID: id})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if repo.deleteCalled != 1 {
		t.Errorf("expected Delete called once, got %d", repo.deleteCalled)
	}
	if repo.deletedID != id {
		t.Errorf("deleted ID mismatch: got %v want %v", repo.deletedID, id)
	}
}
