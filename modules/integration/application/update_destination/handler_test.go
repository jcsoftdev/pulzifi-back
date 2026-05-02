package updatedestination_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	updatedestination "github.com/jcsoftdev/pulzifi-back/modules/integration/application/update_destination"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

type stubDestinationRepo struct {
	getByIDFn    func(ctx context.Context, id uuid.UUID) (*entities.Destination, error)
	updateCalled int
	updateFn     func(ctx context.Context, d *entities.Destination) error
}

func (r *stubDestinationRepo) GetByID(ctx context.Context, id uuid.UUID) (*entities.Destination, error) {
	if r.getByIDFn != nil {
		return r.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (r *stubDestinationRepo) Update(ctx context.Context, d *entities.Destination) error {
	r.updateCalled++
	if r.updateFn != nil {
		return r.updateFn(ctx, d)
	}
	return nil
}
func (r *stubDestinationRepo) Create(_ context.Context, _ *entities.Destination) error { return nil }
func (r *stubDestinationRepo) Delete(_ context.Context, _ uuid.UUID) error              { return nil }
func (r *stubDestinationRepo) ListByScope(_ context.Context, _ entities.ScopeType, _ uuid.UUID) ([]*entities.Destination, error) {
	return nil, nil
}
func (r *stubDestinationRepo) ResolveForEvent(_ context.Context, _ string, _ uuid.UUID, _, _ *uuid.UUID) ([]*entities.Destination, error) {
	return nil, nil
}
func (r *stubDestinationRepo) DisableByIntegrationID(_ context.Context, _ uuid.UUID) error {
	return nil
}

func existingDestination(id uuid.UUID) *entities.Destination {
	return &entities.Destination{
		ID:          id,
		ServiceType: "slack",
		ScopeType:   entities.ScopeWorkspace,
		ScopeID:     uuid.New(),
		Target:      map[string]any{"channel_id": "C000"},
		Events:      []string{"change.detected"},
		Enabled:     true,
	}
}

func TestUpdate_HappyPath(t *testing.T) {
	id := uuid.New()
	original := existingDestination(id)

	newTarget := map[string]any{"channel_id": "C999"}
	newEvents := []string{"change.detected", "alert.created"}
	disabled := false

	var captured *entities.Destination
	repo := &stubDestinationRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*entities.Destination, error) {
			return original, nil
		},
		updateFn: func(_ context.Context, d *entities.Destination) error {
			captured = d
			return nil
		},
	}

	h := updatedestination.NewHandler(repo)
	resp, err := h.Handle(context.Background(), updatedestination.Request{
		ID:      id,
		Target:  newTarget,
		Events:  newEvents,
		Enabled: &disabled,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if repo.updateCalled != 1 {
		t.Errorf("expected Update called once, got %d", repo.updateCalled)
	}
	if captured.Target["channel_id"] != "C999" {
		t.Errorf("target not updated: %v", captured.Target)
	}
	if len(captured.Events) != 2 {
		t.Errorf("events not updated: %v", captured.Events)
	}
	if captured.Enabled {
		t.Error("Enabled should be false")
	}
	if resp.Destination.ID != id {
		t.Errorf("response destination ID mismatch")
	}
}

func TestUpdate_NotFound(t *testing.T) {
	repo := &stubDestinationRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*entities.Destination, error) {
			return nil, nil
		},
	}
	h := updatedestination.NewHandler(repo)
	_, err := h.Handle(context.Background(), updatedestination.Request{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error for not found destination")
	}
	if repo.updateCalled != 0 {
		t.Errorf("expected Update NOT called, got %d", repo.updateCalled)
	}
}

func TestUpdate_EmptyEvents_Rejected(t *testing.T) {
	id := uuid.New()
	repo := &stubDestinationRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*entities.Destination, error) {
			return existingDestination(id), nil
		},
	}
	h := updatedestination.NewHandler(repo)
	_, err := h.Handle(context.Background(), updatedestination.Request{
		ID:     id,
		Events: []string{}, // empty slice = patch with empty → should fail
	})
	if err == nil {
		t.Fatal("expected error for empty events patch")
	}
	if repo.updateCalled != 0 {
		t.Errorf("expected Update NOT called, got %d", repo.updateCalled)
	}
}

func TestUpdate_NilFieldsAreNoOps(t *testing.T) {
	id := uuid.New()
	original := existingDestination(id)

	repo := &stubDestinationRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*entities.Destination, error) {
			return original, nil
		},
	}

	h := updatedestination.NewHandler(repo)
	// Pass only ID — all patch fields are nil/zero-value
	resp, err := h.Handle(context.Background(), updatedestination.Request{ID: id})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	// repo.Update IS still called (we persist unchanged)
	if repo.updateCalled != 1 {
		t.Errorf("expected Update called once, got %d", repo.updateCalled)
	}
	// Values are unchanged
	if resp.Destination.Target["channel_id"] != "C000" {
		t.Errorf("target should be unchanged: %v", resp.Destination.Target)
	}
	if len(resp.Destination.Events) != 1 || resp.Destination.Events[0] != "change.detected" {
		t.Errorf("events should be unchanged: %v", resp.Destination.Events)
	}
	if !resp.Destination.Enabled {
		t.Error("Enabled should remain true")
	}
}
