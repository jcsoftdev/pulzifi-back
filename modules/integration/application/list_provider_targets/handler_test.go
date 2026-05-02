package listprovidertargets_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	listprovidertargets "github.com/jcsoftdev/pulzifi-back/modules/integration/application/list_provider_targets"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
)

// --- stub IntegrationRepository ---

type stubIntegrationRepo struct {
	getByOrgAndServiceFn func(ctx context.Context, orgID uuid.UUID, serviceType string) (*entities.Integration, error)
}

func (r *stubIntegrationRepo) GetByOrgAndService(ctx context.Context, orgID uuid.UUID, serviceType string) (*entities.Integration, error) {
	if r.getByOrgAndServiceFn != nil {
		return r.getByOrgAndServiceFn(ctx, orgID, serviceType)
	}
	return nil, nil
}
func (r *stubIntegrationRepo) Create(_ context.Context, _ *entities.Integration) error { return nil }
func (r *stubIntegrationRepo) Update(_ context.Context, _ *entities.Integration) error { return nil }
func (r *stubIntegrationRepo) GetByID(_ context.Context, _ uuid.UUID) (*entities.Integration, error) {
	return nil, nil
}
func (r *stubIntegrationRepo) ListByOrg(_ context.Context, _ uuid.UUID) ([]*entities.Integration, error) {
	return nil, nil
}
func (r *stubIntegrationRepo) SoftDelete(_ context.Context, _ uuid.UUID) error { return nil }

// --- stub ProviderClient ---

type stubProviderClient struct {
	serviceType    string
	listTargetsFn  func(ctx context.Context, integ *entities.Integration) ([]entities.Target, error)
}

func (c *stubProviderClient) ServiceType() string { return c.serviceType }
func (c *stubProviderClient) OAuthAuthorizeURL(_, _ string) (string, error) {
	return "", nil
}
func (c *stubProviderClient) HandleCallback(_ context.Context, _, _ string) (*entities.OAuthResult, error) {
	return nil, nil
}
func (c *stubProviderClient) RefreshAccessToken(_ context.Context, _ string) (*entities.OAuthResult, error) {
	return nil, nil
}
func (c *stubProviderClient) ListTargets(ctx context.Context, integ *entities.Integration) ([]entities.Target, error) {
	if c.listTargetsFn != nil {
		return c.listTargetsFn(ctx, integ)
	}
	return nil, nil
}
func (c *stubProviderClient) Send(_ context.Context, _ *entities.Integration, _ *entities.Destination, _ *entities.NotificationPayload) (*entities.DeliveryResult, error) {
	return nil, nil
}

// --- stub ProviderRegistry ---

type stubRegistry struct {
	clients map[string]services.ProviderClient
}

func (r *stubRegistry) Get(t string) (services.ProviderClient, bool) {
	c, ok := r.clients[t]
	return c, ok
}

// --- tests ---

func TestListTargets_HappyPath(t *testing.T) {
	orgID := uuid.New()
	integ := &entities.Integration{ID: uuid.New(), OrgID: orgID, ServiceType: "slack"}

	want := []entities.Target{
		{ID: "C001", Name: "general"},
		{ID: "C002", Name: "alerts"},
	}

	repo := &stubIntegrationRepo{
		getByOrgAndServiceFn: func(_ context.Context, _ uuid.UUID, _ string) (*entities.Integration, error) {
			return integ, nil
		},
	}
	client := &stubProviderClient{
		serviceType: "slack",
		listTargetsFn: func(_ context.Context, _ *entities.Integration) ([]entities.Target, error) {
			return want, nil
		},
	}
	registry := &stubRegistry{clients: map[string]services.ProviderClient{"slack": client}}

	h := listprovidertargets.NewHandler(repo, registry)
	resp, err := h.Handle(context.Background(), listprovidertargets.Request{
		OrgID:       orgID,
		ServiceType: "slack",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(resp.Targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(resp.Targets))
	}
}

func TestListTargets_NoIntegration(t *testing.T) {
	repo := &stubIntegrationRepo{
		getByOrgAndServiceFn: func(_ context.Context, _ uuid.UUID, _ string) (*entities.Integration, error) {
			return nil, nil
		},
	}
	registry := &stubRegistry{clients: map[string]services.ProviderClient{}}

	h := listprovidertargets.NewHandler(repo, registry)
	_, err := h.Handle(context.Background(), listprovidertargets.Request{
		OrgID:       uuid.New(),
		ServiceType: "slack",
	})
	if err == nil {
		t.Fatal("expected error for no active integration")
	}
}

func TestListTargets_UnknownProvider(t *testing.T) {
	orgID := uuid.New()
	integ := &entities.Integration{ID: uuid.New(), OrgID: orgID, ServiceType: "slack"}

	repo := &stubIntegrationRepo{
		getByOrgAndServiceFn: func(_ context.Context, _ uuid.UUID, _ string) (*entities.Integration, error) {
			return integ, nil
		},
	}
	// Registry has no entry for "slack"
	registry := &stubRegistry{clients: map[string]services.ProviderClient{}}

	h := listprovidertargets.NewHandler(repo, registry)
	_, err := h.Handle(context.Background(), listprovidertargets.Request{
		OrgID:       orgID,
		ServiceType: "slack",
	})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestListTargets_ClientError(t *testing.T) {
	orgID := uuid.New()
	integ := &entities.Integration{ID: uuid.New(), OrgID: orgID, ServiceType: "slack"}

	repo := &stubIntegrationRepo{
		getByOrgAndServiceFn: func(_ context.Context, _ uuid.UUID, _ string) (*entities.Integration, error) {
			return integ, nil
		},
	}
	client := &stubProviderClient{
		serviceType: "slack",
		listTargetsFn: func(_ context.Context, _ *entities.Integration) ([]entities.Target, error) {
			return nil, errors.New("slack API error")
		},
	}
	registry := &stubRegistry{clients: map[string]services.ProviderClient{"slack": client}}

	h := listprovidertargets.NewHandler(repo, registry)
	_, err := h.Handle(context.Background(), listprovidertargets.Request{
		OrgID:       orgID,
		ServiceType: "slack",
	})
	if err == nil {
		t.Fatal("expected error from client.ListTargets")
	}
}
