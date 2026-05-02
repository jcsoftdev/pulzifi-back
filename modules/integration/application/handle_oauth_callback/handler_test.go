package handleoauthcallback_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	handleoauthcallback "github.com/jcsoftdev/pulzifi-back/modules/integration/application/handle_oauth_callback"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
	intoauth "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/oauth"
)

// ── in-memory IntegrationRepository ──────────────────────────────────────────

type inMemRepo struct {
	rows        map[string]*entities.Integration // key: orgID+":"+serviceType
	createCalls int
	updateCalls int
}

// Compile-time interface check.
var _ repositories.IntegrationRepository = (*inMemRepo)(nil)

func newInMemRepo() *inMemRepo {
	return &inMemRepo{rows: make(map[string]*entities.Integration)}
}

func repoKey(orgID uuid.UUID, serviceType string) string {
	return orgID.String() + ":" + serviceType
}

func (r *inMemRepo) Create(_ context.Context, i *entities.Integration) error {
	r.createCalls++
	cp := *i
	r.rows[repoKey(i.OrgID, i.ServiceType)] = &cp
	return nil
}

func (r *inMemRepo) Update(_ context.Context, i *entities.Integration) error {
	r.updateCalls++
	cp := *i
	r.rows[repoKey(i.OrgID, i.ServiceType)] = &cp
	return nil
}

func (r *inMemRepo) GetByID(_ context.Context, id uuid.UUID) (*entities.Integration, error) {
	for _, v := range r.rows {
		if v.ID == id {
			cp := *v
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *inMemRepo) GetByOrgAndService(_ context.Context, orgID uuid.UUID, serviceType string) (*entities.Integration, error) {
	v, ok := r.rows[repoKey(orgID, serviceType)]
	if !ok {
		return nil, nil
	}
	cp := *v
	return &cp, nil
}

func (r *inMemRepo) ListByOrg(_ context.Context, orgID uuid.UUID) ([]*entities.Integration, error) {
	var out []*entities.Integration
	for _, v := range r.rows {
		if v.OrgID == orgID {
			cp := *v
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *inMemRepo) SoftDelete(_ context.Context, id uuid.UUID) error {
	for k, v := range r.rows {
		if v.ID == id {
			delete(r.rows, k)
			return nil
		}
	}
	return nil
}

// ── stub ProviderClient ───────────────────────────────────────────────────────

type stubClient struct {
	serviceType      string
	handleCallbackFn func(ctx context.Context, code, redirectURI string) (*entities.OAuthResult, error)
	callbackCalls    int
}

func (s *stubClient) ServiceType() string { return s.serviceType }
func (s *stubClient) OAuthAuthorizeURL(state, redirect string) (string, error) {
	return "", nil
}
func (s *stubClient) HandleCallback(ctx context.Context, code, redirectURI string) (*entities.OAuthResult, error) {
	s.callbackCalls++
	return s.handleCallbackFn(ctx, code, redirectURI)
}
func (s *stubClient) RefreshAccessToken(context.Context, string) (*entities.OAuthResult, error) {
	return nil, nil
}
func (s *stubClient) ListTargets(context.Context, *entities.Integration) ([]entities.Target, error) {
	return nil, nil
}
func (s *stubClient) Send(context.Context, *entities.Integration, *entities.Destination, *entities.NotificationPayload) (*entities.DeliveryResult, error) {
	return nil, nil
}

type stubRegistry struct{ clients map[string]services.ProviderClient }

func (r *stubRegistry) Get(t string) (services.ProviderClient, bool) {
	c, ok := r.clients[t]
	return c, ok
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newSigner(t *testing.T, ttl time.Duration) *intoauth.StateSigner {
	t.Helper()
	return intoauth.NewStateSigner(bytes.Repeat([]byte{0x42}, 32), ttl)
}

func validState(t *testing.T, signer *intoauth.StateSigner, provider string, orgID, userID uuid.UUID) string {
	t.Helper()
	tok, err := signer.Sign(intoauth.StateClaims{
		Provider:    provider,
		Tenant:      "acme",
		OrgID:       orgID,
		UserID:      userID,
		ReturnPath:  "/settings/integrations",
		RedirectURI: "https://app.example.com/oauth/callback",
	})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	return tok
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestHandleCallback_NewIntegration_HappyPath(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	signer := newSigner(t, time.Hour)
	state := validState(t, signer, "slack", orgID, userID)

	client := &stubClient{
		serviceType: "slack",
		handleCallbackFn: func(_ context.Context, _, _ string) (*entities.OAuthResult, error) {
			return &entities.OAuthResult{
				AccessToken:  "xoxb",
				ProviderMeta: map[string]any{"team_id": "T1"},
			}, nil
		},
	}
	repo := newInMemRepo()
	h := handleoauthcallback.NewHandler(
		&stubRegistry{clients: map[string]services.ProviderClient{"slack": client}},
		signer,
		repo,
	)

	resp, err := h.Handle(context.Background(), handleoauthcallback.Request{
		Provider: "slack",
		Code:     "code123",
		State:    state,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	// Response fields populated from state.
	if resp.Tenant != "acme" {
		t.Errorf("Tenant = %q, want acme", resp.Tenant)
	}
	if resp.ReturnPath != "/settings/integrations" {
		t.Errorf("ReturnPath = %q, want /settings/integrations", resp.ReturnPath)
	}
	if resp.Provider != "slack" {
		t.Errorf("Provider = %q, want slack", resp.Provider)
	}

	// Exactly one integration written via Create.
	if repo.createCalls != 1 {
		t.Errorf("createCalls = %d, want 1", repo.createCalls)
	}
	if repo.updateCalls != 0 {
		t.Errorf("updateCalls = %d, want 0", repo.updateCalls)
	}

	// Integration stored with correct fields.
	integ, _ := repo.GetByOrgAndService(context.Background(), orgID, "slack")
	if integ == nil {
		t.Fatal("integration not found in repo")
	}
	if integ.AccessToken != "xoxb" {
		t.Errorf("AccessToken = %q, want xoxb", integ.AccessToken)
	}
	if integ.Status != entities.IntegrationActive {
		t.Errorf("Status = %q, want active", integ.Status)
	}
	if integ.OrgID != orgID {
		t.Errorf("OrgID mismatch")
	}
	if integ.CreatedBy != userID {
		t.Errorf("CreatedBy mismatch")
	}
}

func TestHandleCallback_ReplacesExisting(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	existingID := uuid.New()
	signer := newSigner(t, time.Hour)
	state := validState(t, signer, "slack", orgID, userID)

	// Pre-seed the repo with an existing integration.
	repo := newInMemRepo()
	repo.rows[repoKey(orgID, "slack")] = &entities.Integration{
		ID:          existingID,
		OrgID:       orgID,
		ServiceType: "slack",
		Status:      entities.IntegrationActive,
		AccessToken: "old",
		CreatedBy:   userID,
	}

	client := &stubClient{
		serviceType: "slack",
		handleCallbackFn: func(_ context.Context, _, _ string) (*entities.OAuthResult, error) {
			return &entities.OAuthResult{AccessToken: "new"}, nil
		},
	}
	h := handleoauthcallback.NewHandler(
		&stubRegistry{clients: map[string]services.ProviderClient{"slack": client}},
		signer,
		repo,
	)

	_, err := h.Handle(context.Background(), handleoauthcallback.Request{
		Provider: "slack",
		Code:     "code456",
		State:    state,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	// Update called once, Create not called.
	if repo.updateCalls != 1 {
		t.Errorf("updateCalls = %d, want 1", repo.updateCalls)
	}
	if repo.createCalls != 0 {
		t.Errorf("createCalls = %d, want 0 (should not create new row)", repo.createCalls)
	}

	// Still exactly one row for this org+service with same ID and updated token.
	integ, _ := repo.GetByOrgAndService(context.Background(), orgID, "slack")
	if integ == nil {
		t.Fatal("integration not found after update")
	}
	if integ.ID != existingID {
		t.Errorf("ID changed: got %v, want %v", integ.ID, existingID)
	}
	if integ.AccessToken != "new" {
		t.Errorf("AccessToken = %q, want new", integ.AccessToken)
	}

	// Only one row total.
	all, _ := repo.ListByOrg(context.Background(), orgID)
	if len(all) != 1 {
		t.Errorf("len(rows) = %d, want 1", len(all))
	}
}

func TestHandleCallback_ExpiredState(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	// Sign with negative TTL so the token is already expired.
	signer := newSigner(t, -time.Second)
	state := validState(t, signer, "slack", orgID, userID)

	// Use a fresh signer with normal TTL to verify — verification should fail on expiry.
	verifier := newSigner(t, time.Hour)
	repo := newInMemRepo()
	client := &stubClient{
		serviceType: "slack",
		handleCallbackFn: func(_ context.Context, _, _ string) (*entities.OAuthResult, error) {
			return &entities.OAuthResult{AccessToken: "tok"}, nil
		},
	}
	h := handleoauthcallback.NewHandler(
		&stubRegistry{clients: map[string]services.ProviderClient{"slack": client}},
		verifier,
		repo,
	)

	_, err := h.Handle(context.Background(), handleoauthcallback.Request{
		Provider: "slack",
		Code:     "code",
		State:    state,
	})
	if err == nil {
		t.Fatal("expected error for expired state, got nil")
	}
	if !strings.Contains(err.Error(), "invalid or expired state") {
		t.Errorf("error = %q, want to contain 'invalid or expired state'", err.Error())
	}
	if repo.createCalls != 0 || repo.updateCalls != 0 {
		t.Errorf("no DB writes expected, got create=%d update=%d", repo.createCalls, repo.updateCalls)
	}
}

func TestHandleCallback_TamperedState(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	signer := newSigner(t, time.Hour)
	state := validState(t, signer, "slack", orgID, userID)

	// Tamper the first character of the header part.
	tampered := state
	if tampered[0] == 'A' {
		tampered = "B" + tampered[1:]
	} else {
		tampered = "A" + tampered[1:]
	}

	repo := newInMemRepo()
	client := &stubClient{
		serviceType: "slack",
		handleCallbackFn: func(_ context.Context, _, _ string) (*entities.OAuthResult, error) {
			return &entities.OAuthResult{AccessToken: "tok"}, nil
		},
	}
	h := handleoauthcallback.NewHandler(
		&stubRegistry{clients: map[string]services.ProviderClient{"slack": client}},
		signer,
		repo,
	)

	_, err := h.Handle(context.Background(), handleoauthcallback.Request{
		Provider: "slack",
		Code:     "code",
		State:    tampered,
	})
	if err == nil {
		t.Fatal("expected error for tampered state, got nil")
	}
	// Accept either "invalid or expired state" (wrapper) or "bad signature" (inner).
	if !strings.Contains(err.Error(), "invalid or expired state") && !strings.Contains(err.Error(), "bad signature") {
		t.Errorf("error = %q, want to contain 'invalid or expired state' or 'bad signature'", err.Error())
	}
	if repo.createCalls != 0 || repo.updateCalls != 0 {
		t.Errorf("no DB writes expected, got create=%d update=%d", repo.createCalls, repo.updateCalls)
	}
}

func TestHandleCallback_ProviderMismatch(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	signer := newSigner(t, time.Hour)
	// Sign state for "slack" but request claims "discord".
	state := validState(t, signer, "slack", orgID, userID)

	repo := newInMemRepo()
	discordClient := &stubClient{
		serviceType: "discord",
		handleCallbackFn: func(_ context.Context, _, _ string) (*entities.OAuthResult, error) {
			return &entities.OAuthResult{AccessToken: "tok"}, nil
		},
	}
	h := handleoauthcallback.NewHandler(
		&stubRegistry{clients: map[string]services.ProviderClient{
			"slack":   &stubClient{serviceType: "slack", handleCallbackFn: func(_ context.Context, _, _ string) (*entities.OAuthResult, error) { return nil, errors.New("should not be called") }},
			"discord": discordClient,
		}},
		signer,
		repo,
	)

	_, err := h.Handle(context.Background(), handleoauthcallback.Request{
		Provider: "discord",
		Code:     "code",
		State:    state,
	})
	if err == nil {
		t.Fatal("expected provider mismatch error, got nil")
	}
	if !strings.Contains(err.Error(), "provider mismatch") {
		t.Errorf("error = %q, want to contain 'provider mismatch'", err.Error())
	}
	// HandleCallback on the discord stub must NOT have been called.
	if discordClient.callbackCalls != 0 {
		t.Errorf("discord HandleCallback called %d times, want 0", discordClient.callbackCalls)
	}
	if repo.createCalls != 0 || repo.updateCalls != 0 {
		t.Errorf("no DB writes expected, got create=%d update=%d", repo.createCalls, repo.updateCalls)
	}
}

func TestHandleCallback_CallbackExchangeFailure(t *testing.T) {
	orgID := uuid.New()
	userID := uuid.New()
	signer := newSigner(t, time.Hour)
	state := validState(t, signer, "slack", orgID, userID)

	exchangeErr := errors.New("slack oauth: invalid_code")
	client := &stubClient{
		serviceType: "slack",
		handleCallbackFn: func(_ context.Context, _, _ string) (*entities.OAuthResult, error) {
			return nil, exchangeErr
		},
	}
	repo := newInMemRepo()
	h := handleoauthcallback.NewHandler(
		&stubRegistry{clients: map[string]services.ProviderClient{"slack": client}},
		signer,
		repo,
	)

	_, err := h.Handle(context.Background(), handleoauthcallback.Request{
		Provider: "slack",
		Code:     "badcode",
		State:    state,
	})
	if err == nil {
		t.Fatal("expected error from callback exchange, got nil")
	}
	if !strings.Contains(err.Error(), "invalid_code") {
		t.Errorf("error = %q, want to contain 'invalid_code'", err.Error())
	}
	if repo.createCalls != 0 || repo.updateCalls != 0 {
		t.Errorf("no DB writes expected after exchange failure, got create=%d update=%d", repo.createCalls, repo.updateCalls)
	}
}
