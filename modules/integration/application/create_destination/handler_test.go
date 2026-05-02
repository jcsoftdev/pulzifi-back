package createdestination_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	createdestination "github.com/jcsoftdev/pulzifi-back/modules/integration/application/create_destination"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

type stubDestinationRepo struct {
	createCalled int
	createFn     func(ctx context.Context, d *entities.Destination) error
}

func (r *stubDestinationRepo) Create(ctx context.Context, d *entities.Destination) error {
	r.createCalled++
	if r.createFn != nil {
		return r.createFn(ctx, d)
	}
	return nil
}
func (r *stubDestinationRepo) Update(_ context.Context, _ *entities.Destination) error { return nil }
func (r *stubDestinationRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.Destination, error) {
	return nil, nil
}
func (r *stubDestinationRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }
func (r *stubDestinationRepo) ListByScope(_ context.Context, _ entities.ScopeType, _ uuid.UUID) ([]*entities.Destination, error) {
	return nil, nil
}
func (r *stubDestinationRepo) ResolveForEvent(_ context.Context, _ string, _ uuid.UUID, _, _ *uuid.UUID) ([]*entities.Destination, error) {
	return nil, nil
}
func (r *stubDestinationRepo) DisableByIntegrationID(_ context.Context, _ uuid.UUID) error {
	return nil
}

func validRequest() createdestination.Request {
	return createdestination.Request{
		ServiceType: "slack",
		ScopeType:   entities.ScopeWorkspace,
		ScopeID:     uuid.New(),
		Target:      map[string]any{"channel_id": "C123"},
		Events:      []string{"change.detected"},
	}
}

func TestCreate_HappyPath(t *testing.T) {
	repo := &stubDestinationRepo{}
	h := createdestination.NewHandler(repo)

	resp, err := h.Handle(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if repo.createCalled != 1 {
		t.Errorf("expected repo.Create called once, got %d", repo.createCalled)
	}
	if resp.Destination.ID == uuid.Nil {
		t.Error("expected Destination.ID to be set")
	}
	if !resp.Destination.Enabled {
		t.Error("expected Destination.Enabled to be true")
	}
}

func TestCreate_MissingServiceType(t *testing.T) {
	repo := &stubDestinationRepo{}
	h := createdestination.NewHandler(repo)

	req := validRequest()
	req.ServiceType = ""
	_, err := h.Handle(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for missing service_type")
	}
	if repo.createCalled != 0 {
		t.Errorf("expected repo.Create NOT called, got %d", repo.createCalled)
	}
}

func TestCreate_InvalidScopeType(t *testing.T) {
	repo := &stubDestinationRepo{}
	h := createdestination.NewHandler(repo)

	req := validRequest()
	req.ScopeType = "banana"
	_, err := h.Handle(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for invalid scope_type")
	}
	if repo.createCalled != 0 {
		t.Errorf("expected repo.Create NOT called, got %d", repo.createCalled)
	}
}

func TestCreate_EmptyEvents(t *testing.T) {
	repo := &stubDestinationRepo{}
	h := createdestination.NewHandler(repo)

	req := validRequest()
	req.Events = []string{}
	_, err := h.Handle(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty events")
	}
	if repo.createCalled != 0 {
		t.Errorf("expected repo.Create NOT called, got %d", repo.createCalled)
	}
}

func TestCreate_NilTarget(t *testing.T) {
	repo := &stubDestinationRepo{}
	h := createdestination.NewHandler(repo)

	req := validRequest()
	req.Target = nil
	_, err := h.Handle(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for nil target")
	}
	if repo.createCalled != 0 {
		t.Errorf("expected repo.Create NOT called, got %d", repo.createCalled)
	}
}

func TestCreate_NilScopeID(t *testing.T) {
	repo := &stubDestinationRepo{}
	h := createdestination.NewHandler(repo)

	req := validRequest()
	req.ScopeID = uuid.Nil
	_, err := h.Handle(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for nil scope_id")
	}
	if repo.createCalled != 0 {
		t.Errorf("expected repo.Create NOT called, got %d", repo.createCalled)
	}
}
