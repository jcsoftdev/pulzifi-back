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

func (r *inMemDelRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.Delivery, error) {
	return nil, nil
}

func (r *inMemDelRepo) ClaimPending(_ context.Context, limit int, now time.Time) ([]*entities.Delivery, error) {
	return nil, nil
}

func (r *inMemDelRepo) MarkDelivered(_ context.Context, id uuid.UUID, code int, bodySnip string) error {
	return nil
}

func (r *inMemDelRepo) MarkFailed(_ context.Context, id uuid.UUID, code *int, body, errMsg string, nextAttempt time.Time, _ []entities.Attempt) error {
	return nil
}

func (r *inMemDelRepo) MarkDead(_ context.Context, id uuid.UUID, errMsg string, _ []entities.Attempt) error {
	return nil
}

func (r *inMemDelRepo) ListByDestination(_ context.Context, destinationID uuid.UUID, limit, offset int) ([]*entities.Delivery, error) {
	return nil, nil
}

func (r *inMemDelRepo) Retry(_ context.Context, id uuid.UUID) error { return nil }

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
// ChannelEntitlement stub
// ---------------------------------------------------------------------------

type stubEntitlement struct {
	emailOnly bool
	err       error
}

func (s *stubEntitlement) IsPaid(_ context.Context, _ uuid.UUID) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return !s.emailOnly, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// stubFallback is an in-memory FallbackRecipientResolver.
type stubFallback struct {
	emails []string
	err    error
	calls  int
}

func (s *stubFallback) WorkspaceMemberEmails(_ context.Context, _ string, _ uuid.UUID) ([]string, error) {
	s.calls++
	return s.emails, s.err
}

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

// When no email destination resolves for the event, the handler seeds a
// workspace-scoped email destination from the workspace's members and delivers to it.
func TestDispatch_NoEmailDestination_SeedsWorkspaceMembersFallback(t *testing.T) {
	orgID := uuid.New()
	wsID := uuid.New()

	destRepo := &inMemDestRepo{} // nothing configured
	delRepo := &inMemDelRepo{}
	h := NewHandlerWithEntitlement(&stubFactory{destRepo, delRepo}, &stubOrgGuard{isActive: true}, &stubEntitlement{emailOnly: true})
	h.SetFallbackResolver(&stubFallback{emails: []string{"a@x.com", "b@x.com"}})

	ev := makeEvent(orgID, &wsID, nil)
	if err := h.Handle(context.Background(), ev); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if len(destRepo.dests) != 1 {
		t.Fatalf("seeded destinations = %d, want 1", len(destRepo.dests))
	}
	d := destRepo.dests[0]
	if d.ServiceType != "email" || d.ScopeType != entities.ScopeWorkspace || d.ScopeID != wsID || !d.Enabled {
		t.Errorf("seeded destination wrong: %+v", d)
	}
	if !slices.Contains(d.Events, "change.detected") {
		t.Errorf("seeded destination events = %v, want change.detected", d.Events)
	}
	if len(delRepo.created) != 1 {
		t.Fatalf("deliveries = %d, want 1", len(delRepo.created))
	}
	if delRepo.created[0].DestinationID != d.ID {
		t.Errorf("delivery destination = %v, want seeded %v", delRepo.created[0].DestinationID, d.ID)
	}
}

// When an email destination already exists, the fallback must NOT fire (no seeding,
// no duplicate) — the configured destination decides recipients.
func TestDispatch_EmailDestinationExists_NoFallbackSeed(t *testing.T) {
	orgID := uuid.New()
	wsID := uuid.New()

	existing := newDest("email", entities.ScopeOrg, orgID, []string{"change.detected"})
	destRepo := &inMemDestRepo{dests: []*entities.Destination{existing}}
	delRepo := &inMemDelRepo{}
	fb := &stubFallback{emails: []string{"a@x.com"}}
	h := NewHandlerWithEntitlement(&stubFactory{destRepo, delRepo}, &stubOrgGuard{isActive: true}, &stubEntitlement{emailOnly: true})
	h.SetFallbackResolver(fb)

	ev := makeEvent(orgID, &wsID, nil)
	if err := h.Handle(context.Background(), ev); err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if fb.calls != 0 {
		t.Errorf("fallback invoked %d times, want 0 (email destination exists)", fb.calls)
	}
	if len(destRepo.dests) != 1 {
		t.Errorf("destinations = %d, want 1 (no seeding)", len(destRepo.dests))
	}
	if len(delRepo.created) != 1 || delRepo.created[0].DestinationID != existing.ID {
		t.Errorf("expected 1 delivery to the existing destination")
	}
}

// No members to fall back to → no seed, no delivery.
func TestDispatch_NoEmailDest_NoMembers_NoSeed(t *testing.T) {
	orgID := uuid.New()
	wsID := uuid.New()

	destRepo := &inMemDestRepo{}
	delRepo := &inMemDelRepo{}
	h := NewHandlerWithEntitlement(&stubFactory{destRepo, delRepo}, &stubOrgGuard{isActive: true}, &stubEntitlement{emailOnly: true})
	h.SetFallbackResolver(&stubFallback{emails: nil})

	ev := makeEvent(orgID, &wsID, nil)
	if err := h.Handle(context.Background(), ev); err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(destRepo.dests) != 0 || len(delRepo.created) != 0 {
		t.Errorf("expected no seed/delivery; dests=%d deliveries=%d", len(destRepo.dests), len(delRepo.created))
	}
}

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

// ---------------------------------------------------------------------------
// ChannelEntitlement / plan-tiered tests
// ---------------------------------------------------------------------------

// TestDispatch_FreeOrg_OnlyEmailDelivered asserts that when the entitlement
// gate reports email-only, non-email destinations are skipped and only the
// email destination receives a Delivery row.
func TestDispatch_FreeOrg_OnlyEmailDelivered(t *testing.T) {
	orgID := uuid.New()

	slackDest := newDest("slack", entities.ScopeOrg, orgID, []string{"change.detected"})
	discordDest := newDest("discord", entities.ScopeOrg, orgID, []string{"change.detected"})
	emailDest := newDest("email", entities.ScopeOrg, orgID, []string{"change.detected"})

	destRepo := &inMemDestRepo{dests: []*entities.Destination{slackDest, discordDest, emailDest}}
	delRepo := &inMemDelRepo{}
	h := NewHandlerWithEntitlement(
		&stubFactory{destRepo, delRepo},
		&stubOrgGuard{isActive: true},
		&stubEntitlement{emailOnly: true},
	)

	ev := makeEvent(orgID, nil, nil)
	if err := h.Handle(context.Background(), ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(delRepo.created) != 1 {
		t.Fatalf("expected 1 delivery (email only), got %d", len(delRepo.created))
	}
	if delRepo.created[0].DestinationID != emailDest.ID {
		t.Errorf("expected email destination, got destination %v", delRepo.created[0].DestinationID)
	}
}

// TestDispatch_PaidOrg_AllChannelsDelivered asserts that when the org is on a
// paid plan, all destinations receive Delivery rows regardless of ServiceType.
func TestDispatch_PaidOrg_AllChannelsDelivered(t *testing.T) {
	orgID := uuid.New()

	slackDest := newDest("slack", entities.ScopeOrg, orgID, []string{"change.detected"})
	discordDest := newDest("discord", entities.ScopeOrg, orgID, []string{"change.detected"})
	emailDest := newDest("email", entities.ScopeOrg, orgID, []string{"change.detected"})

	destRepo := &inMemDestRepo{dests: []*entities.Destination{slackDest, discordDest, emailDest}}
	delRepo := &inMemDelRepo{}
	h := NewHandlerWithEntitlement(
		&stubFactory{destRepo, delRepo},
		&stubOrgGuard{isActive: true},
		&stubEntitlement{emailOnly: false},
	)

	ev := makeEvent(orgID, nil, nil)
	if err := h.Handle(context.Background(), ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(delRepo.created) != 3 {
		t.Fatalf("expected 3 deliveries (all channels), got %d", len(delRepo.created))
	}
}

// TestDispatch_NilEntitlement_AllChannelsDelivered asserts that when no
// ChannelEntitlement is injected (nil), all destinations pass through
// unchanged — preserving backward-compatible behaviour for call sites that
// don't supply the gate.
func TestDispatch_NilEntitlement_AllChannelsDelivered(t *testing.T) {
	orgID := uuid.New()

	slackDest := newDest("slack", entities.ScopeOrg, orgID, []string{"change.detected"})
	emailDest := newDest("email", entities.ScopeOrg, orgID, []string{"change.detected"})

	destRepo := &inMemDestRepo{dests: []*entities.Destination{slackDest, emailDest}}
	delRepo := &inMemDelRepo{}
	// NewHandler (no entitlement) must behave as "all paid".
	h := NewHandler(&stubFactory{destRepo, delRepo}, &stubOrgGuard{isActive: true})

	ev := makeEvent(orgID, nil, nil)
	if err := h.Handle(context.Background(), ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(delRepo.created) != 2 {
		t.Fatalf("expected 2 deliveries (nil entitlement = all pass), got %d", len(delRepo.created))
	}
}

// TestDispatch_EntitlementError_FailOpenEmailOnly asserts that when the
// ChannelEntitlement check returns an error, the handler fails open to
// email-only (does NOT silently allow all channels) and returns no error
// to the caller (alert email still goes out; paid channels are skipped as a
// conservative safe-default).
func TestDispatch_EntitlementError_FailOpenEmailOnly(t *testing.T) {
	orgID := uuid.New()

	slackDest := newDest("slack", entities.ScopeOrg, orgID, []string{"change.detected"})
	emailDest := newDest("email", entities.ScopeOrg, orgID, []string{"change.detected"})

	destRepo := &inMemDestRepo{dests: []*entities.Destination{slackDest, emailDest}}
	delRepo := &inMemDelRepo{}
	h := NewHandlerWithEntitlement(
		&stubFactory{destRepo, delRepo},
		&stubOrgGuard{isActive: true},
		&stubEntitlement{err: errors.New("db unavailable")},
	)

	ev := makeEvent(orgID, nil, nil)
	// Must not return an error to the caller — email path must still succeed.
	if err := h.Handle(context.Background(), ev); err != nil {
		t.Fatalf("expected no error on entitlement failure (fail-open), got: %v", err)
	}

	if len(delRepo.created) != 1 {
		t.Fatalf("expected 1 delivery (email only on error), got %d", len(delRepo.created))
	}
	if delRepo.created[0].DestinationID != emailDest.ID {
		t.Errorf("expected email destination on entitlement error")
	}
}
