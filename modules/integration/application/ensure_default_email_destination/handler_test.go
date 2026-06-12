package ensuredefaultemaildestination_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	ensuredefault "github.com/jcsoftdev/pulzifi-back/modules/integration/application/ensure_default_email_destination"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

type stubDestinationRepo struct {
	existing     []*entities.Destination
	listErr      error
	createErr    error
	createCalled int
	created      *entities.Destination
}

func (r *stubDestinationRepo) Create(_ context.Context, d *entities.Destination) error {
	r.createCalled++
	r.created = d
	return r.createErr
}
func (r *stubDestinationRepo) Update(_ context.Context, _ *entities.Destination) error { return nil }
func (r *stubDestinationRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.Destination, error) {
	return nil, nil
}
func (r *stubDestinationRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }
func (r *stubDestinationRepo) ListByScope(_ context.Context, _ entities.ScopeType, _ uuid.UUID) ([]*entities.Destination, error) {
	return r.existing, r.listErr
}
func (r *stubDestinationRepo) ResolveForEvent(_ context.Context, _ string, _ uuid.UUID, _, _ *uuid.UUID) ([]*entities.Destination, error) {
	return nil, nil
}
func (r *stubDestinationRepo) DisableByIntegrationID(_ context.Context, _ uuid.UUID) error {
	return nil
}

func TestEnsure_CreatesDefaultWhenNoEmailDestination(t *testing.T) {
	repo := &stubDestinationRepo{}
	h := ensuredefault.NewHandler(repo)
	orgID := uuid.New()

	resp, err := h.Handle(context.Background(), ensuredefault.Request{
		OrgID:      orgID,
		OwnerEmail: "owner@acme.com",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !resp.Created {
		t.Fatal("expected Created=true")
	}
	if repo.createCalled != 1 {
		t.Fatalf("expected one Create call, got %d", repo.createCalled)
	}
	d := repo.created
	if d.ServiceType != "email" {
		t.Errorf("ServiceType = %q, want email", d.ServiceType)
	}
	if d.ScopeType != entities.ScopeOrg || d.ScopeID != orgID {
		t.Errorf("scope = %s/%s, want org/%s", d.ScopeType, d.ScopeID, orgID)
	}
	emails, _ := d.Target["emails"].([]any)
	if len(emails) != 1 || emails[0] != "owner@acme.com" {
		t.Errorf("Target[emails] = %v, want [owner@acme.com]", d.Target["emails"])
	}
	if len(d.Events) != 2 {
		t.Errorf("Events = %v, want change.detected + alert.created", d.Events)
	}
	if !d.Enabled {
		t.Error("expected Enabled=true")
	}
	if d.IntegrationID != nil {
		t.Error("expected nil IntegrationID for email (no-auth provider)")
	}
}

func TestEnsure_NoopWhenEmailDestinationExists(t *testing.T) {
	repo := &stubDestinationRepo{
		existing: []*entities.Destination{{ServiceType: "email"}},
	}
	h := ensuredefault.NewHandler(repo)

	resp, err := h.Handle(context.Background(), ensuredefault.Request{
		OrgID:      uuid.New(),
		OwnerEmail: "owner@acme.com",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if resp.Created {
		t.Error("expected Created=false")
	}
	if repo.createCalled != 0 {
		t.Errorf("expected no Create call, got %d", repo.createCalled)
	}
}

func TestEnsure_IgnoresOtherServiceTypes(t *testing.T) {
	repo := &stubDestinationRepo{
		existing: []*entities.Destination{{ServiceType: "slack"}},
	}
	h := ensuredefault.NewHandler(repo)

	resp, err := h.Handle(context.Background(), ensuredefault.Request{
		OrgID:      uuid.New(),
		OwnerEmail: "owner@acme.com",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !resp.Created {
		t.Error("expected Created=true when only non-email destinations exist")
	}
}

func TestEnsure_NoopWhenOwnerEmailEmpty(t *testing.T) {
	repo := &stubDestinationRepo{}
	h := ensuredefault.NewHandler(repo)

	resp, err := h.Handle(context.Background(), ensuredefault.Request{
		OrgID:      uuid.New(),
		OwnerEmail: "",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if resp.Created || repo.createCalled != 0 {
		t.Error("expected no-op when owner email is empty")
	}
}

func TestEnsure_NoopWhenOrgIDNil(t *testing.T) {
	repo := &stubDestinationRepo{}
	h := ensuredefault.NewHandler(repo)

	resp, err := h.Handle(context.Background(), ensuredefault.Request{
		OrgID:      uuid.Nil,
		OwnerEmail: "owner@acme.com",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if resp.Created || repo.createCalled != 0 {
		t.Error("expected no-op when org id is nil")
	}
}

func TestEnsure_PropagatesListError(t *testing.T) {
	repo := &stubDestinationRepo{listErr: errors.New("boom")}
	h := ensuredefault.NewHandler(repo)

	_, err := h.Handle(context.Background(), ensuredefault.Request{
		OrgID:      uuid.New(),
		OwnerEmail: "owner@acme.com",
	})
	if err == nil {
		t.Fatal("expected error from ListByScope")
	}
}

func TestEnsure_PropagatesCreateError(t *testing.T) {
	repo := &stubDestinationRepo{createErr: errors.New("boom")}
	h := ensuredefault.NewHandler(repo)

	_, err := h.Handle(context.Background(), ensuredefault.Request{
		OrgID:      uuid.New(),
		OwnerEmail: "owner@acme.com",
	})
	if err == nil {
		t.Fatal("expected error from Create")
	}
}
