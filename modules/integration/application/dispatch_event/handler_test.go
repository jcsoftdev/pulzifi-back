package dispatchevent

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/shared/eventbus"
)

// ---------------------------------------------------------------------------
// In-memory DestinationRepository
// ---------------------------------------------------------------------------

var _ repositories.DestinationRepository = (*inMemDestRepo)(nil)

type inMemDestRepo struct {
	dests []*entities.Destination
}

func (r *inMemDestRepo) Create(_ context.Context, d *entities.Destination) error {
	r.dests = append(r.dests, d)
	return nil
}

func (r *inMemDestRepo) Update(_ context.Context, d *entities.Destination) error { return nil }

func (r *inMemDestRepo) GetByID(_ context.Context, id uuid.UUID) (*entities.Destination, error) {
	return nil, nil
}

func (r *inMemDestRepo) Delete(_ context.Context, id uuid.UUID) error { return nil }

func (r *inMemDestRepo) ListByScope(_ context.Context, scope entities.ScopeType, scopeID uuid.UUID) ([]*entities.Destination, error) {
	return nil, nil
}

func (r *inMemDestRepo) DisableByIntegrationID(_ context.Context, integrationID uuid.UUID) error {
	return nil
}

// ResolveForEvent implements the scope-override logic in Go (mirrors SQL behaviour).
// Rules:
//  1. Include destination if enabled && eventType ∈ destination.Events && scope matches.
//  2. Assign specificity: page=3, workspace=2, org=1.
//  3. Group by ServiceType; keep only rows at the maximum specificity in each group.
func (r *inMemDestRepo) ResolveForEvent(
	ctx context.Context,
	eventType string,
	orgID uuid.UUID,
	wsID, pageID *uuid.UUID,
) ([]*entities.Destination, error) {
	type candidate struct {
		dest        *entities.Destination
		specificity int
	}

	// Step 1: filter
	var candidates []candidate
	for _, d := range r.dests {
		if !d.Enabled {
			continue
		}
		if !slices.Contains(d.Events, eventType) {
			continue
		}
		var spec int
		switch d.ScopeType {
		case entities.ScopeOrg:
			if d.ScopeID != orgID {
				continue
			}
			spec = 1
		case entities.ScopeWorkspace:
			if wsID == nil || d.ScopeID != *wsID {
				continue
			}
			spec = 2
		case entities.ScopePage:
			if pageID == nil || d.ScopeID != *pageID {
				continue
			}
			spec = 3
		default:
			continue
		}
		candidates = append(candidates, candidate{dest: d, specificity: spec})
	}

	// Step 2: per-ServiceType max specificity
	maxSpec := map[string]int{}
	for _, c := range candidates {
		if c.specificity > maxSpec[c.dest.ServiceType] {
			maxSpec[c.dest.ServiceType] = c.specificity
		}
	}

	// Step 3: keep winners
	var result []*entities.Destination
	for _, c := range candidates {
		if c.specificity == maxSpec[c.dest.ServiceType] {
			result = append(result, c.dest)
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// In-memory DeliveryRepository
// ---------------------------------------------------------------------------

var _ repositories.DeliveryRepository = (*inMemDelRepo)(nil)

type inMemDelRepo struct {
	created    []*entities.Delivery
	failOnCall int // if > 0, fail on that ordinal Create call (1-based)
	calls      int
}

func (r *inMemDelRepo) Create(_ context.Context, d *entities.Delivery) error {
	r.calls++
	if r.failOnCall > 0 && r.calls == r.failOnCall {
		return errors.New("simulated persistence error")
	}
	r.created = append(r.created, d)
	return nil
}

func (r *inMemDelRepo) ClaimPending(_ context.Context, limit int, now time.Time) ([]*entities.Delivery, error) {
	return nil, nil
}

func (r *inMemDelRepo) MarkDelivered(_ context.Context, id uuid.UUID, code int, bodySnip string) error {
	return nil
}

func (r *inMemDelRepo) MarkFailed(_ context.Context, id uuid.UUID, code *int, body, errMsg string, nextAttempt time.Time) error {
	return nil
}

func (r *inMemDelRepo) MarkDead(_ context.Context, id uuid.UUID, errMsg string) error { return nil }

func (r *inMemDelRepo) ListByDestination(_ context.Context, destinationID uuid.UUID, limit, offset int) ([]*entities.Delivery, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// RepoFactory stub
// ---------------------------------------------------------------------------

type stubFactory struct {
	destRepo *inMemDestRepo
	delRepo  *inMemDelRepo
}

func (f *stubFactory) DestinationRepoForTenant(_ string) repositories.DestinationRepository {
	return f.destRepo
}

func (f *stubFactory) DeliveryRepoForTenant(_ string) repositories.DeliveryRepository {
	return f.delRepo
}

// ---------------------------------------------------------------------------
// OrgGuard stub
// ---------------------------------------------------------------------------

type stubOrgGuard struct {
	isActive bool
}

func (g *stubOrgGuard) IsActive(_ context.Context, _ uuid.UUID) (bool, error) {
	return g.isActive, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newDest(serviceType string, scope entities.ScopeType, scopeID uuid.UUID, events []string) *entities.Destination {
	return &entities.Destination{
		ID:          uuid.New(),
		ServiceType: serviceType,
		ScopeType:   scope,
		ScopeID:     scopeID,
		Events:      events,
		Enabled:     true,
	}
}

func makeEvent(orgID uuid.UUID, wsID, pageID *uuid.UUID) eventbus.DomainEvent {
	return eventbus.DomainEvent{
		ID:          uuid.New(),
		Type:        "change.detected",
		OccurredAt:  time.Now().UTC(),
		OrgID:       orgID,
		WorkspaceID: wsID,
		PageID:      pageID,
		Tenant:      "test-tenant",
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestDispatch_OrgOnly_OneDelivery(t *testing.T) {
	orgID := uuid.New()
	dest := newDest("slack", entities.ScopeOrg, orgID, []string{"change.detected"})

	destRepo := &inMemDestRepo{dests: []*entities.Destination{dest}}
	delRepo := &inMemDelRepo{}
	h := NewHandler(&stubFactory{destRepo, delRepo}, &stubOrgGuard{isActive: true})

	ev := makeEvent(orgID, nil, nil)
	if err := h.Handle(context.Background(), ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(delRepo.created) != 1 {
		t.Fatalf("expected 1 delivery, got %d", len(delRepo.created))
	}
	d := delRepo.created[0]
	if d.DestinationID != dest.ID {
		t.Errorf("DestinationID mismatch: want %v got %v", dest.ID, d.DestinationID)
	}
	if d.EventType != "change.detected" {
		t.Errorf("EventType mismatch: want change.detected got %s", d.EventType)
	}
	if d.Status != entities.DeliveryPending {
		t.Errorf("Status mismatch: want pending got %s", d.Status)
	}
}

func TestDispatch_WorkspaceOverridesOrg(t *testing.T) {
	orgID := uuid.New()
	wsID := uuid.New()

	orgDest := newDest("slack", entities.ScopeOrg, orgID, []string{"change.detected"})
	wsDest := newDest("slack", entities.ScopeWorkspace, wsID, []string{"change.detected"})

	destRepo := &inMemDestRepo{dests: []*entities.Destination{orgDest, wsDest}}
	delRepo := &inMemDelRepo{}
	h := NewHandler(&stubFactory{destRepo, delRepo}, &stubOrgGuard{isActive: true})

	ev := makeEvent(orgID, &wsID, nil)
	if err := h.Handle(context.Background(), ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(delRepo.created) != 1 {
		t.Fatalf("expected 1 delivery (workspace wins), got %d", len(delRepo.created))
	}
	if delRepo.created[0].DestinationID != wsDest.ID {
		t.Errorf("expected workspace dest to win")
	}
}

func TestDispatch_PageOverridesAll(t *testing.T) {
	orgID := uuid.New()
	wsID := uuid.New()
	pageID := uuid.New()

	orgDest := newDest("slack", entities.ScopeOrg, orgID, []string{"change.detected"})
	wsDest := newDest("slack", entities.ScopeWorkspace, wsID, []string{"change.detected"})
	pageDest := newDest("slack", entities.ScopePage, pageID, []string{"change.detected"})

	destRepo := &inMemDestRepo{dests: []*entities.Destination{orgDest, wsDest, pageDest}}
	delRepo := &inMemDelRepo{}
	h := NewHandler(&stubFactory{destRepo, delRepo}, &stubOrgGuard{isActive: true})

	ev := makeEvent(orgID, &wsID, &pageID)
	if err := h.Handle(context.Background(), ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(delRepo.created) != 1 {
		t.Fatalf("expected 1 delivery (page wins), got %d", len(delRepo.created))
	}
	if delRepo.created[0].DestinationID != pageDest.ID {
		t.Errorf("expected page dest to win")
	}
}

func TestDispatch_DifferentServicesBothFire(t *testing.T) {
	orgID := uuid.New()
	wsID := uuid.New()

	orgSlack := newDest("slack", entities.ScopeOrg, orgID, []string{"change.detected"})
	wsEmail := newDest("email", entities.ScopeWorkspace, wsID, []string{"change.detected"})

	destRepo := &inMemDestRepo{dests: []*entities.Destination{orgSlack, wsEmail}}
	delRepo := &inMemDelRepo{}
	h := NewHandler(&stubFactory{destRepo, delRepo}, &stubOrgGuard{isActive: true})

	ev := makeEvent(orgID, &wsID, nil)
	if err := h.Handle(context.Background(), ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(delRepo.created) != 2 {
		t.Fatalf("expected 2 deliveries (different services), got %d", len(delRepo.created))
	}
}

func TestDispatch_MultipleAtWinningScope(t *testing.T) {
	orgID := uuid.New()

	dest1 := newDest("slack", entities.ScopeOrg, orgID, []string{"change.detected"})
	dest2 := newDest("slack", entities.ScopeOrg, orgID, []string{"change.detected"})

	destRepo := &inMemDestRepo{dests: []*entities.Destination{dest1, dest2}}
	delRepo := &inMemDelRepo{}
	h := NewHandler(&stubFactory{destRepo, delRepo}, &stubOrgGuard{isActive: true})

	ev := makeEvent(orgID, nil, nil)
	if err := h.Handle(context.Background(), ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(delRepo.created) != 2 {
		t.Fatalf("expected 2 deliveries, got %d", len(delRepo.created))
	}
}

func TestDispatch_OrgInactive_NoDeliveries(t *testing.T) {
	orgID := uuid.New()
	dest := newDest("slack", entities.ScopeOrg, orgID, []string{"change.detected"})

	destRepo := &inMemDestRepo{dests: []*entities.Destination{dest}}
	delRepo := &inMemDelRepo{}
	h := NewHandler(&stubFactory{destRepo, delRepo}, &stubOrgGuard{isActive: false})

	ev := makeEvent(orgID, nil, nil)
	if err := h.Handle(context.Background(), ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(delRepo.created) != 0 {
		t.Fatalf("expected 0 deliveries for inactive org, got %d", len(delRepo.created))
	}
}

func TestDispatch_EventDataPreserved(t *testing.T) {
	orgID := uuid.New()
	dest := newDest("slack", entities.ScopeOrg, orgID, []string{"change.detected"})

	destRepo := &inMemDestRepo{dests: []*entities.Destination{dest}}
	delRepo := &inMemDelRepo{}
	h := NewHandler(&stubFactory{destRepo, delRepo}, &stubOrgGuard{isActive: true})

	ev := makeEvent(orgID, nil, nil)
	ev.Data = json.RawMessage(`{"foo":"bar"}`)

	if err := h.Handle(context.Background(), ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(delRepo.created) != 1 {
		t.Fatalf("expected 1 delivery, got %d", len(delRepo.created))
	}
	payload := delRepo.created[0].EventPayload
	if payload["foo"] != "bar" {
		t.Errorf("expected foo=bar in payload, got %v", payload)
	}
}

func TestDispatch_NilEventData(t *testing.T) {
	orgID := uuid.New()
	dest := newDest("slack", entities.ScopeOrg, orgID, []string{"change.detected"})

	destRepo := &inMemDestRepo{dests: []*entities.Destination{dest}}
	delRepo := &inMemDelRepo{}
	h := NewHandler(&stubFactory{destRepo, delRepo}, &stubOrgGuard{isActive: true})

	ev := makeEvent(orgID, nil, nil)
	ev.Data = nil // explicitly nil

	if err := h.Handle(context.Background(), ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(delRepo.created) != 1 {
		t.Fatalf("expected 1 delivery, got %d", len(delRepo.created))
	}
	payload := delRepo.created[0].EventPayload
	if payload == nil {
		t.Errorf("EventPayload must not be nil (JSONB column is NOT NULL)")
	}
}

func TestDispatch_PersistenceErrorStops(t *testing.T) {
	orgID := uuid.New()

	dest1 := newDest("slack", entities.ScopeOrg, orgID, []string{"change.detected"})
	dest2 := newDest("email", entities.ScopeOrg, orgID, []string{"change.detected"})

	destRepo := &inMemDestRepo{dests: []*entities.Destination{dest1, dest2}}
	delRepo := &inMemDelRepo{failOnCall: 2} // succeed on call 1, fail on call 2
	h := NewHandler(&stubFactory{destRepo, delRepo}, &stubOrgGuard{isActive: true})

	ev := makeEvent(orgID, nil, nil)
	err := h.Handle(context.Background(), ev)
	if err == nil {
		t.Fatal("expected error from persistence, got nil")
	}

	if len(delRepo.created) != 1 {
		t.Fatalf("expected exactly 1 delivery persisted before error, got %d", len(delRepo.created))
	}
}
