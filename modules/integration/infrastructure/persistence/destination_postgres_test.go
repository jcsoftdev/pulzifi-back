//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/persistence"
)

const testTenant = "jcsoftdev_inc"

// newDestRepo returns a DestinationRepository wired to the test tenant.
func newDestRepo(t *testing.T) interface {
	Create(ctx context.Context, d *entities.Destination) error
	Update(ctx context.Context, d *entities.Destination) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Destination, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByScope(ctx context.Context, scope entities.ScopeType, scopeID uuid.UUID) ([]*entities.Destination, error)
	ResolveForEvent(ctx context.Context, eventType string, orgID uuid.UUID, workspaceID, pageID *uuid.UUID) ([]*entities.Destination, error)
	DisableByIntegrationID(ctx context.Context, integrationID uuid.UUID) error
} {
	t.Helper()
	db := openTestDB(t)
	t.Cleanup(func() { db.Close() })
	return persistence.NewDestinationPostgresRepository(db, testTenant)
}

// insertDest is a helper that creates a Destination and registers cleanup.
func insertDest(t *testing.T, ctx context.Context, repo interface {
	Create(ctx context.Context, d *entities.Destination) error
	Delete(ctx context.Context, id uuid.UUID) error
}, d *entities.Destination) {
	t.Helper()
	if err := repo.Create(ctx, d); err != nil {
		t.Fatalf("insertDest: Create: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.Delete(ctx, d.ID); err != nil {
			t.Logf("insertDest cleanup: Delete %s: %v", d.ID, err)
		}
	})
}

// --------------------------------------------------------------------------
// TestDestinationRepo_RoundTrip: Create → GetByID → Update → ListByScope → Delete
// --------------------------------------------------------------------------

func TestDestinationRepo_RoundTrip(t *testing.T) {
	ctx := context.Background()
	repo := newDestRepo(t)

	orgID := uuid.New()

	d := &entities.Destination{
		ID:          uuid.New(),
		ServiceType: "slack",
		ScopeType:   entities.ScopeOrg,
		ScopeID:     orgID,
		Target:      map[string]any{"webhook_url": "https://hooks.slack.com/test"},
		Events:      []string{"change.detected"},
		Enabled:     true,
	}

	// Create
	if err := repo.Create(ctx, d); err != nil {
		t.Fatalf("Create: %v", err)
	}
	t.Cleanup(func() {
		if err := repo.Delete(ctx, d.ID); err != nil {
			t.Logf("cleanup Delete %s: %v", d.ID, err)
		}
	})

	// GetByID
	got, err := repo.GetByID(ctx, d.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil {
		t.Fatal("GetByID returned nil")
	}
	if got.ID != d.ID {
		t.Errorf("ID mismatch: want %s got %s", d.ID, got.ID)
	}
	if got.ServiceType != "slack" {
		t.Errorf("ServiceType mismatch: got %q", got.ServiceType)
	}
	if got.ScopeType != entities.ScopeOrg {
		t.Errorf("ScopeType mismatch: got %q", got.ScopeType)
	}
	if len(got.Events) != 1 || got.Events[0] != "change.detected" {
		t.Errorf("Events mismatch: got %v", got.Events)
	}
	if got.Target["webhook_url"] != "https://hooks.slack.com/test" {
		t.Errorf("Target mismatch: got %v", got.Target)
	}

	// Update
	got.Target = map[string]any{"webhook_url": "https://hooks.slack.com/updated"}
	got.Events = []string{"change.detected", "alert.created"}
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}

	updated, err := repo.GetByID(ctx, d.ID)
	if err != nil {
		t.Fatalf("GetByID after Update: %v", err)
	}
	if updated.Target["webhook_url"] != "https://hooks.slack.com/updated" {
		t.Errorf("updated Target mismatch: got %v", updated.Target)
	}
	if len(updated.Events) != 2 {
		t.Errorf("updated Events length: want 2 got %d", len(updated.Events))
	}

	// ListByScope
	list, err := repo.ListByScope(ctx, entities.ScopeOrg, orgID)
	if err != nil {
		t.Fatalf("ListByScope: %v", err)
	}
	found := false
	for _, item := range list {
		if item.ID == d.ID {
			found = true
		}
	}
	if !found {
		t.Errorf("ListByScope: created destination not found")
	}

	// Delete
	if err := repo.Delete(ctx, d.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// GetByID after Delete returns nil
	afterDelete, err := repo.GetByID(ctx, d.ID)
	if err != nil {
		t.Fatalf("GetByID after Delete: %v", err)
	}
	if afterDelete != nil {
		t.Errorf("expected nil after Delete, got %+v", afterDelete)
	}
}

// --------------------------------------------------------------------------
// TestDestinationRepo_ResolveForEvent_OrgOnly
// Org-level Slack dest fires for any event in that org (no workspace/page).
// --------------------------------------------------------------------------

func TestDestinationRepo_ResolveForEvent_OrgOnly(t *testing.T) {
	ctx := context.Background()
	repo := newDestRepo(t)

	orgID := uuid.New()

	d := &entities.Destination{
		ID:          uuid.New(),
		ServiceType: "slack",
		ScopeType:   entities.ScopeOrg,
		ScopeID:     orgID,
		Target:      map[string]any{},
		Events:      []string{"change.detected"},
		Enabled:     true,
	}
	insertDest(t, ctx, repo, d)

	results, err := repo.ResolveForEvent(ctx, "change.detected", orgID, nil, nil)
	if err != nil {
		t.Fatalf("ResolveForEvent: %v", err)
	}
	if !containsID(results, d.ID) {
		t.Errorf("expected org dest to fire; got %d results", len(results))
	}
}

// --------------------------------------------------------------------------
// TestDestinationRepo_ResolveForEvent_WorkspaceOverride
// Org Slack + workspace Slack → only workspace fires for events under that workspace.
// --------------------------------------------------------------------------

func TestDestinationRepo_ResolveForEvent_WorkspaceOverride(t *testing.T) {
	ctx := context.Background()
	repo := newDestRepo(t)

	orgID := uuid.New()
	wsID := uuid.New()

	orgDest := &entities.Destination{
		ID:          uuid.New(),
		ServiceType: "slack",
		ScopeType:   entities.ScopeOrg,
		ScopeID:     orgID,
		Target:      map[string]any{},
		Events:      []string{"change.detected"},
		Enabled:     true,
	}
	wsDest := &entities.Destination{
		ID:          uuid.New(),
		ServiceType: "slack",
		ScopeType:   entities.ScopeWorkspace,
		ScopeID:     wsID,
		Target:      map[string]any{},
		Events:      []string{"change.detected"},
		Enabled:     true,
	}
	insertDest(t, ctx, repo, orgDest)
	insertDest(t, ctx, repo, wsDest)

	results, err := repo.ResolveForEvent(ctx, "change.detected", orgID, &wsID, nil)
	if err != nil {
		t.Fatalf("ResolveForEvent: %v", err)
	}
	if containsID(results, orgDest.ID) {
		t.Errorf("org dest should be overridden by workspace dest")
	}
	if !containsID(results, wsDest.ID) {
		t.Errorf("workspace dest should fire")
	}
}

// --------------------------------------------------------------------------
// TestDestinationRepo_ResolveForEvent_PageOverride
// Org + workspace + page Slack → only page fires.
// --------------------------------------------------------------------------

func TestDestinationRepo_ResolveForEvent_PageOverride(t *testing.T) {
	ctx := context.Background()
	repo := newDestRepo(t)

	orgID := uuid.New()
	wsID := uuid.New()
	pageID := uuid.New()

	orgDest := &entities.Destination{
		ID:          uuid.New(),
		ServiceType: "slack",
		ScopeType:   entities.ScopeOrg,
		ScopeID:     orgID,
		Target:      map[string]any{},
		Events:      []string{"change.detected"},
		Enabled:     true,
	}
	wsDest := &entities.Destination{
		ID:          uuid.New(),
		ServiceType: "slack",
		ScopeType:   entities.ScopeWorkspace,
		ScopeID:     wsID,
		Target:      map[string]any{},
		Events:      []string{"change.detected"},
		Enabled:     true,
	}
	pageDest := &entities.Destination{
		ID:          uuid.New(),
		ServiceType: "slack",
		ScopeType:   entities.ScopePage,
		ScopeID:     pageID,
		Target:      map[string]any{},
		Events:      []string{"change.detected"},
		Enabled:     true,
	}
	insertDest(t, ctx, repo, orgDest)
	insertDest(t, ctx, repo, wsDest)
	insertDest(t, ctx, repo, pageDest)

	results, err := repo.ResolveForEvent(ctx, "change.detected", orgID, &wsID, &pageID)
	if err != nil {
		t.Fatalf("ResolveForEvent: %v", err)
	}
	if containsID(results, orgDest.ID) || containsID(results, wsDest.ID) {
		t.Errorf("org/workspace dests should be overridden by page dest")
	}
	if !containsID(results, pageDest.ID) {
		t.Errorf("page dest should fire")
	}
}

// --------------------------------------------------------------------------
// TestDestinationRepo_ResolveForEvent_DifferentServices
// Org Slack + workspace Email → both fire (different service_type).
// --------------------------------------------------------------------------

func TestDestinationRepo_ResolveForEvent_DifferentServices(t *testing.T) {
	ctx := context.Background()
	repo := newDestRepo(t)

	orgID := uuid.New()
	wsID := uuid.New()

	slackDest := &entities.Destination{
		ID:          uuid.New(),
		ServiceType: "slack",
		ScopeType:   entities.ScopeOrg,
		ScopeID:     orgID,
		Target:      map[string]any{},
		Events:      []string{"change.detected"},
		Enabled:     true,
	}
	emailDest := &entities.Destination{
		ID:          uuid.New(),
		ServiceType: "email",
		ScopeType:   entities.ScopeWorkspace,
		ScopeID:     wsID,
		Target:      map[string]any{},
		Events:      []string{"change.detected"},
		Enabled:     true,
	}
	insertDest(t, ctx, repo, slackDest)
	insertDest(t, ctx, repo, emailDest)

	results, err := repo.ResolveForEvent(ctx, "change.detected", orgID, &wsID, nil)
	if err != nil {
		t.Fatalf("ResolveForEvent: %v", err)
	}
	if !containsID(results, slackDest.ID) {
		t.Errorf("org Slack dest should fire (no workspace Slack override)")
	}
	if !containsID(results, emailDest.ID) {
		t.Errorf("workspace Email dest should fire")
	}
}

// --------------------------------------------------------------------------
// TestDestinationRepo_ResolveForEvent_MultipleAtWinningScope
// Two org Slack dests (same scope, same service) → both fire.
// --------------------------------------------------------------------------

func TestDestinationRepo_ResolveForEvent_MultipleAtWinningScope(t *testing.T) {
	ctx := context.Background()
	repo := newDestRepo(t)

	orgID := uuid.New()

	d1 := &entities.Destination{
		ID:          uuid.New(),
		ServiceType: "slack",
		ScopeType:   entities.ScopeOrg,
		ScopeID:     orgID,
		Target:      map[string]any{"ch": "1"},
		Events:      []string{"change.detected"},
		Enabled:     true,
	}
	d2 := &entities.Destination{
		ID:          uuid.New(),
		ServiceType: "slack",
		ScopeType:   entities.ScopeOrg,
		ScopeID:     orgID,
		Target:      map[string]any{"ch": "2"},
		Events:      []string{"change.detected"},
		Enabled:     true,
	}
	insertDest(t, ctx, repo, d1)
	insertDest(t, ctx, repo, d2)

	results, err := repo.ResolveForEvent(ctx, "change.detected", orgID, nil, nil)
	if err != nil {
		t.Fatalf("ResolveForEvent: %v", err)
	}
	if !containsID(results, d1.ID) || !containsID(results, d2.ID) {
		t.Errorf("both org Slack dests should fire; got %d results", len(results))
	}
}

// --------------------------------------------------------------------------
// TestDestinationRepo_DisableByIntegrationID
// Two enabled dests share an integration_id → after Disable, both enabled=false.
// --------------------------------------------------------------------------

func TestDestinationRepo_DisableByIntegrationID(t *testing.T) {
	ctx := context.Background()
	repo := newDestRepo(t)

	integrationID := uuid.New()
	orgID := uuid.New()

	d1 := &entities.Destination{
		ID:            uuid.New(),
		IntegrationID: &integrationID,
		ServiceType:   "slack",
		ScopeType:     entities.ScopeOrg,
		ScopeID:       orgID,
		Target:        map[string]any{},
		Events:        []string{"change.detected"},
		Enabled:       true,
	}
	d2 := &entities.Destination{
		ID:            uuid.New(),
		IntegrationID: &integrationID,
		ServiceType:   "email",
		ScopeType:     entities.ScopeOrg,
		ScopeID:       orgID,
		Target:        map[string]any{},
		Events:        []string{"change.detected"},
		Enabled:       true,
	}
	insertDest(t, ctx, repo, d1)
	insertDest(t, ctx, repo, d2)

	if err := repo.DisableByIntegrationID(ctx, integrationID); err != nil {
		t.Fatalf("DisableByIntegrationID: %v", err)
	}

	got1, err := repo.GetByID(ctx, d1.ID)
	if err != nil {
		t.Fatalf("GetByID d1: %v", err)
	}
	if got1 == nil || got1.Enabled {
		t.Errorf("d1 should be disabled; got %+v", got1)
	}

	got2, err := repo.GetByID(ctx, d2.ID)
	if err != nil {
		t.Fatalf("GetByID d2: %v", err)
	}
	if got2 == nil || got2.Enabled {
		t.Errorf("d2 should be disabled; got %+v", got2)
	}
}

// --------------------------------------------------------------------------
// helpers
// --------------------------------------------------------------------------

func containsID(list []*entities.Destination, id uuid.UUID) bool {
	for _, d := range list {
		if d.ID == id {
			return true
		}
	}
	return false
}
