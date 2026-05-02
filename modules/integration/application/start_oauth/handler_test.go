package startoauth_test

import (
	"bytes"
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/application/start_oauth"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
	intoauth "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/oauth"
)

type stubClient struct {
	serviceType string
	authorizeFn func(state, redirect string) (string, error)
}

func (s *stubClient) ServiceType() string { return s.serviceType }
func (s *stubClient) OAuthAuthorizeURL(state, redirect string) (string, error) {
	return s.authorizeFn(state, redirect)
}
func (s *stubClient) HandleCallback(context.Context, string, string) (*entities.OAuthResult, error) {
	return nil, nil
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

func newSigner(t *testing.T) *intoauth.StateSigner {
	t.Helper()
	return intoauth.NewStateSigner(bytes.Repeat([]byte{0x42}, 32), time.Hour)
}

func TestStartOAuth_HappyPath(t *testing.T) {
	var capturedState, capturedRedirect string
	client := &stubClient{
		serviceType: "slack",
		authorizeFn: func(state, redirect string) (string, error) {
			capturedState, capturedRedirect = state, redirect
			return "https://slack.com/oauth/v2/authorize?state=" + url.QueryEscape(state), nil
		},
	}
	signer := newSigner(t)
	h := startoauth.NewHandler(&stubRegistry{clients: map[string]services.ProviderClient{"slack": client}}, signer)

	resp, err := h.Handle(context.Background(), startoauth.Request{
		Provider:    "slack",
		Tenant:      "acme",
		OrgID:       uuid.New(),
		UserID:      uuid.New(),
		ReturnPath:  "/settings/integrations",
		RedirectURI: "https://app.example.com/oauth/callback",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !strings.Contains(resp.AuthorizeURL, "slack.com") {
		t.Errorf("AuthorizeURL missing slack.com: %s", resp.AuthorizeURL)
	}
	if capturedState == "" {
		t.Error("client did not receive state")
	}
	if capturedRedirect != "https://app.example.com/oauth/callback" {
		t.Errorf("redirect mismatch: %s", capturedRedirect)
	}
	claims, err := signer.Verify(capturedState)
	if err != nil {
		t.Fatalf("state did not verify: %v", err)
	}
	if claims.Tenant != "acme" || claims.Provider != "slack" {
		t.Errorf("claims mismatch: %+v", claims)
	}
}

func TestStartOAuth_UnknownProvider(t *testing.T) {
	h := startoauth.NewHandler(&stubRegistry{clients: map[string]services.ProviderClient{}}, newSigner(t))
	if _, err := h.Handle(context.Background(), startoauth.Request{Provider: "nonesuch"}); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}
