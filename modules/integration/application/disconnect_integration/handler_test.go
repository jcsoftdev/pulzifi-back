package disconnectintegration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/application/disconnect_integration"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
)

type stubIntRepo struct {
	softDeleted []uuid.UUID
	softErr     error
}

func (s *stubIntRepo) Create(context.Context, *entities.Integration) error { return nil }
func (s *stubIntRepo) Update(context.Context, *entities.Integration) error { return nil }
func (s *stubIntRepo) GetByID(context.Context, uuid.UUID) (*entities.Integration, error) {
	return nil, nil
}
func (s *stubIntRepo) GetByOrgAndService(context.Context, uuid.UUID, string) (*entities.Integration, error) {
	return nil, nil
}
func (s *stubIntRepo) ListByOrg(context.Context, uuid.UUID) ([]*entities.Integration, error) {
	return nil, nil
}
func (s *stubIntRepo) SoftDelete(_ context.Context, id uuid.UUID) error {
	if s.softErr != nil {
		return s.softErr
	}
	s.softDeleted = append(s.softDeleted, id)
	return nil
}

type stubDestRepo struct {
	disabled  []uuid.UUID
	disableErr error
}

func (s *stubDestRepo) Create(context.Context, *entities.Destination) error { return nil }
func (s *stubDestRepo) Update(context.Context, *entities.Destination) error { return nil }
func (s *stubDestRepo) GetByID(context.Context, uuid.UUID) (*entities.Destination, error) {
	return nil, nil
}
func (s *stubDestRepo) Delete(context.Context, uuid.UUID) error { return nil }
func (s *stubDestRepo) ListByScope(context.Context, entities.ScopeType, uuid.UUID) ([]*entities.Destination, error) {
	return nil, nil
}
func (s *stubDestRepo) ResolveForEvent(context.Context, string, uuid.UUID, *uuid.UUID, *uuid.UUID) ([]*entities.Destination, error) {
	return nil, nil
}
func (s *stubDestRepo) DisableByIntegrationID(_ context.Context, id uuid.UUID) error {
	if s.disableErr != nil {
		return s.disableErr
	}
	s.disabled = append(s.disabled, id)
	return nil
}

var (
	_ repositories.IntegrationRepository = (*stubIntRepo)(nil)
	_ repositories.DestinationRepository = (*stubDestRepo)(nil)
)

func TestDisconnect_HappyPath(t *testing.T) {
	ir := &stubIntRepo{}
	dr := &stubDestRepo{}
	h := disconnectintegration.NewHandler(ir, dr)
	id := uuid.New()
	if _, err := h.Handle(context.Background(), disconnectintegration.Request{IntegrationID: id}); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(ir.softDeleted) != 1 || ir.softDeleted[0] != id {
		t.Errorf("SoftDelete not called with id: %+v", ir.softDeleted)
	}
	if len(dr.disabled) != 1 || dr.disabled[0] != id {
		t.Errorf("DisableByIntegrationID not called with id: %+v", dr.disabled)
	}
}

func TestDisconnect_MissingID(t *testing.T) {
	h := disconnectintegration.NewHandler(&stubIntRepo{}, &stubDestRepo{})
	if _, err := h.Handle(context.Background(), disconnectintegration.Request{}); err == nil {
		t.Fatal("expected error for missing integration_id")
	}
}

func TestDisconnect_SoftDeleteError_Aborts(t *testing.T) {
	ir := &stubIntRepo{softErr: errors.New("boom")}
	dr := &stubDestRepo{}
	h := disconnectintegration.NewHandler(ir, dr)
	if _, err := h.Handle(context.Background(), disconnectintegration.Request{IntegrationID: uuid.New()}); err == nil {
		t.Fatal("expected error")
	}
	if len(dr.disabled) != 0 {
		t.Errorf("DisableByIntegrationID should not run after SoftDelete failure")
	}
}

func TestDisconnect_DisableError_Surfaces(t *testing.T) {
	ir := &stubIntRepo{}
	dr := &stubDestRepo{disableErr: errors.New("boom")}
	h := disconnectintegration.NewHandler(ir, dr)
	if _, err := h.Handle(context.Background(), disconnectintegration.Request{IntegrationID: uuid.New()}); err == nil {
		t.Fatal("expected error")
	}
}
