package startoauth_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/application/start_oauth"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/services"
)

// ── local StateSignerPort implementation ──────────────────────────────────────
// Avoids importing modules/integration/infrastructure/oauth from the application
// test layer (which violates the application→infrastructure boundary).
// Token format mirrors the infrastructure implementation: base64url(json{claims}) + "." + base64url(hmac_sha256(key, header))

type localStateClaims struct {
	Provider    string    `json:"p"`
	Tenant      string    `json:"t"`
	OrgID       uuid.UUID `json:"o"`
	UserID      uuid.UUID `json:"u"`
	ReturnPath  string    `json:"r"`
	RedirectURI string    `json:"d"`
	Nonce       string    `json:"n"`
	ExpiresAt   int64     `json:"e"`
	ReturnHost  string    `json:"h"`
}

type localStateSigner struct {
	key []byte
	ttl time.Duration
}

func newLocalStateSigner(key []byte, ttl time.Duration) services.StateSignerPort {
	return &localStateSigner{key: key, ttl: ttl}
}

func (s *localStateSigner) Sign(c services.StateClaims) (string, error) {
	lc := localStateClaims{
		Provider:    c.Provider,
		Tenant:      c.Tenant,
		OrgID:       c.OrgID,
		UserID:      c.UserID,
		ReturnPath:  c.ReturnPath,
		RedirectURI: c.RedirectURI,
		Nonce:       uuid.NewString(),
		ExpiresAt:   time.Now().Add(s.ttl).Unix(),
		ReturnHost:  c.ReturnHost,
	}
	body, err := json.Marshal(lc)
	if err != nil {
		return "", err
	}
	head := base64.RawURLEncoding.EncodeToString(body)
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(head))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return head + "." + sig, nil
}

func (s *localStateSigner) Verify(token string) (*services.StateClaims, error) {
	dot := strings.IndexByte(token, '.')
	if dot <= 0 {
		return nil, errors.New("malformed state")
	}
	head, sig := token[:dot], token[dot+1:]
	if sig == "" {
		return nil, errors.New("malformed state")
	}
	expected := hmac.New(sha256.New, s.key)
	expected.Write([]byte(head))
	expSig := base64.RawURLEncoding.EncodeToString(expected.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expSig)) {
		return nil, errors.New("bad signature")
	}
	body, err := base64.RawURLEncoding.DecodeString(head)
	if err != nil {
		return nil, err
	}
	var lc localStateClaims
	if err := json.Unmarshal(body, &lc); err != nil {
		return nil, err
	}
	if time.Now().Unix() > lc.ExpiresAt {
		return nil, errors.New("state expired")
	}
	return &services.StateClaims{
		Provider:    lc.Provider,
		Tenant:      lc.Tenant,
		OrgID:       lc.OrgID,
		UserID:      lc.UserID,
		ReturnPath:  lc.ReturnPath,
		RedirectURI: lc.RedirectURI,
		ReturnHost:  lc.ReturnHost,
	}, nil
}

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

func newSigner(t *testing.T) services.StateSignerPort {
	t.Helper()
	return newLocalStateSigner(bytes.Repeat([]byte{0x42}, 32), time.Hour)
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
		ReturnHost:  "http://acme.pulzifi.local:3000",
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
	if claims.ReturnHost != "http://acme.pulzifi.local:3000" {
		t.Errorf("ReturnHost in claims = %q, want http://acme.pulzifi.local:3000", claims.ReturnHost)
	}
}

func TestStartOAuth_UnknownProvider(t *testing.T) {
	h := startoauth.NewHandler(&stubRegistry{clients: map[string]services.ProviderClient{}}, newSigner(t))
	if _, err := h.Handle(context.Background(), startoauth.Request{Provider: "nonesuch"}); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}
