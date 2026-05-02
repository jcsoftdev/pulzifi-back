package listdestinations_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	listdestinations "github.com/jcsoftdev/pulzifi-back/modules/integration/application/list_destinations"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

// stubDestinationRepo implements only the methods needed for list_destinations tests.
type stubDestinationRepo struct {
	listFn func(ctx context.Context, scope entities.ScopeType, scopeID uuid.UUID) ([]*entities.Destination, error)
}

func (r *stubDestinationRepo) ListByScope(ctx context.Context, scope entities.ScopeType, scopeID uuid.UUID) ([]*entities.Destination, error) {
	return r.listFn(ctx, scope, scopeID)
}
func (r *stubDestinationRepo) Create(ctx context.Context, d *entities.Destination) error {
	return nil
}
func (r *stubDestinationRepo) Update(ctx context.Context, d *entities.Destination) error {
	return nil
}
func (r *stubDestinationRepo) GetByID(ctx context.Context, id uuid.UUID) (*entities.Destination, error) {
	return nil, nil
}
func (r *stubDestinationRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (r *stubDestinationRepo) ResolveForEvent(ctx context.Context, eventType string, orgID uuid.UUID, workspaceID, pageID *uuid.UUID) ([]*entities.Destination, error) {
	return nil, nil
}
func (r *stubDestinationRepo) DisableByIntegrationID(ctx context.Context, integrationID uuid.UUID) error {
	return nil
}

func TestListDestinations_HappyPath(t *testing.T) {
	scopeID := uuid.New()
	want := []*entities.Destination{
		{ID: uuid.New(), ScopeType: entities.ScopeWorkspace, ScopeID: scopeID},
		{ID: uuid.New(), ScopeType: entities.ScopeWorkspace, ScopeID: scopeID},
	}

	var capturedScope entities.ScopeType
	var capturedID uuid.UUID

	repo := &stubDestinationRepo{
		listFn: func(_ context.Context, scope entities.ScopeType, sid uuid.UUID) ([]*entities.Destination, error) {
			capturedScope = scope
			capturedID = sid
			return want, nil
		},
	}

	h := listdestinations.NewHandler(repo)
	resp, err := h.Handle(context.Background(), listdestinations.Request{
		ScopeType: entities.ScopeWorkspace,
		ScopeID:   scopeID,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if capturedScope != entities.ScopeWorkspace {
		t.Errorf("scope mismatch: got %q want %q", capturedScope, entities.ScopeWorkspace)
	}
	if capturedID != scopeID {
		t.Errorf("scopeID mismatch: got %v want %v", capturedID, scopeID)
	}
	if len(resp.Destinations) != 2 {
		t.Errorf("expected 2 destinations, got %d", len(resp.Destinations))
	}
}

func TestListDestinations_RepoError(t *testing.T) {
	repo := &stubDestinationRepo{
		listFn: func(_ context.Context, _ entities.ScopeType, _ uuid.UUID) ([]*entities.Destination, error) {
			return nil, errors.New("db error")
		},
	}
	h := listdestinations.NewHandler(repo)
	_, err := h.Handle(context.Background(), listdestinations.Request{
		ScopeType: entities.ScopeOrg,
		ScopeID:   uuid.New(),
	})
	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
}
