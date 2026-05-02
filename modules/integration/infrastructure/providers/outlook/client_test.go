package outlookprovider

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

func newTestClient(t *testing.T) *Client {
	t.Helper()
	return New("test-client-id", "test-client-secret", "http://localhost:3000")
}

func TestOAuthAuthorizeURL_IncludesRequiredParams(t *testing.T) {
	c := newTestClient(t)
	rawURL, err := c.OAuthAuthorizeURL("my-state", "https://example.com/callback")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("invalid URL: %v", err)
	}
	q := parsed.Query()
	if got := q.Get("state"); got != "my-state" {
		t.Errorf("state = %q, want %q", got, "my-state")
	}
	if !strings.Contains(q.Get("scope"), "Mail.Send") {
		t.Errorf("scope %q should contain Mail.Send", q.Get("scope"))
	}
	if !strings.Contains(q.Get("scope"), "offline_access") {
		t.Errorf("scope %q should contain offline_access", q.Get("scope"))
	}
	if !strings.Contains(rawURL, "organizations") {
		t.Errorf("URL %q should use /organizations endpoint", rawURL)
	}
}

func TestHandleCallback_Success(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok-ms","refresh_token":"ref-ms","expires_in":3599}`))
	}))
	defer tokenSrv.Close()

	meSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"mail":"user@company.com","userPrincipalName":"user@company.com"}`))
	}))
	defer meSrv.Close()

	origToken, origUserinfo := TokenURL, UserinfoURL
	TokenURL, UserinfoURL = tokenSrv.URL, meSrv.URL
	t.Cleanup(func() { TokenURL, UserinfoURL = origToken, origUserinfo })

	c := newTestClient(t)
	result, err := c.HandleCallback(t.Context(), "code-abc", "https://example.com/callback")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "tok-ms" {
		t.Errorf("AccessToken = %q, want tok-ms", result.AccessToken)
	}
	if email, _ := result.ProviderMeta["email"].(string); email != "user@company.com" {
		t.Errorf("provider_meta.email = %q, want user@company.com", email)
	}
}

func TestHandleCallback_UsesPrincipalNameFallback(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"tok","refresh_token":"ref","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	// mail field is empty — should fall back to userPrincipalName
	meSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"mail":"","userPrincipalName":"upn@tenant.onmicrosoft.com"}`))
	}))
	defer meSrv.Close()

	origToken, origUserinfo := TokenURL, UserinfoURL
	TokenURL, UserinfoURL = tokenSrv.URL, meSrv.URL
	t.Cleanup(func() { TokenURL, UserinfoURL = origToken, origUserinfo })

	c := newTestClient(t)
	result, err := c.HandleCallback(t.Context(), "code", "https://example.com/callback")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	email, _ := result.ProviderMeta["email"].(string)
	if email != "upn@tenant.onmicrosoft.com" {
		t.Errorf("expected UPN fallback, got %q", email)
	}
}

func TestHandleCallback_BothEmailFieldsAbsent_ReturnsError(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"access_token":"tok","refresh_token":"ref","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	meSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"mail":"","userPrincipalName":""}`))
	}))
	defer meSrv.Close()

	origToken, origUserinfo := TokenURL, UserinfoURL
	TokenURL, UserinfoURL = tokenSrv.URL, meSrv.URL
	t.Cleanup(func() { TokenURL, UserinfoURL = origToken, origUserinfo })

	c := newTestClient(t)
	_, err := c.HandleCallback(t.Context(), "code", "https://example.com/callback")
	if err == nil {
		t.Fatal("expected error when both email fields absent")
	}
}

func TestRefreshAccessToken_ReturnsNewRefreshToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		vals, _ := url.ParseQuery(string(body))
		if vals.Get("grant_type") != "refresh_token" {
			t.Errorf("grant_type = %q, want refresh_token", vals.Get("grant_type"))
		}
		_, _ = w.Write([]byte(`{"access_token":"new-tok","refresh_token":"new-ref","expires_in":3599}`))
	}))
	defer srv.Close()

	orig := TokenURL
	TokenURL = srv.URL
	t.Cleanup(func() { TokenURL = orig })

	c := newTestClient(t)
	result, err := c.RefreshAccessToken(t.Context(), "old-ref")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "new-tok" {
		t.Errorf("AccessToken = %q, want new-tok", result.AccessToken)
	}
	if result.RefreshToken != "new-ref" {
		t.Errorf("RefreshToken = %q, want new-ref (Microsoft replaces refresh token)", result.RefreshToken)
	}
}

func TestSend_Success(t *testing.T) {
	var capturedMsg map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "Bearer ms-tok" {
			t.Errorf("Authorization = %q, want Bearer ms-tok", auth)
		}
		_ = json.NewDecoder(r.Body).Decode(&capturedMsg)
		// Graph sendMail returns 202 Accepted on success
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	orig := SendURL
	SendURL = srv.URL
	t.Cleanup(func() { SendURL = orig })

	c := newTestClient(t)
	integ := &entities.Integration{
		ID:           uuid.New(),
		AccessToken:  "ms-tok",
		ProviderMeta: map[string]any{"email": "sender@company.com"},
	}
	dest := &entities.Destination{
		ID:     uuid.New(),
		Target: map[string]any{"emails": []any{"recipient@example.com"}},
	}
	payload := &entities.NotificationPayload{
		Title:   "Alert",
		Body:    "Something changed",
		PageURL: "https://example.com",
	}

	result, err := c.Send(t.Context(), integ, dest, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Code != http.StatusOK {
		t.Errorf("Code = %d, want %d", result.Code, http.StatusOK)
	}

	// Verify from address is set to provider_meta.email
	msg, _ := capturedMsg["message"].(map[string]any)
	if msg == nil {
		t.Fatal("message field missing in request body")
	}
	from, _ := msg["from"].(map[string]any)
	emailAddr, _ := from["emailAddress"].(map[string]any)
	if addr, _ := emailAddr["address"].(string); addr != "sender@company.com" {
		t.Errorf("from address = %q, want sender@company.com", addr)
	}
}

func TestSend_NilInteg_ReturnsError(t *testing.T) {
	c := newTestClient(t)
	dest := &entities.Destination{Target: map[string]any{"emails": []any{"r@example.com"}}}
	_, err := c.Send(t.Context(), nil, dest, &entities.NotificationPayload{})
	if err == nil {
		t.Fatal("expected error for nil integ")
	}
}

func TestSend_Unauthorized_ReturnsCode401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"code":"InvalidAuthenticationToken"}}`))
	}))
	defer srv.Close()

	orig := SendURL
	SendURL = srv.URL
	t.Cleanup(func() { SendURL = orig })

	c := newTestClient(t)
	integ := &entities.Integration{
		AccessToken:  "bad",
		ProviderMeta: map[string]any{"email": "s@company.com"},
	}
	dest := &entities.Destination{Target: map[string]any{"emails": []any{"r@example.com"}}}
	result, err := c.Send(t.Context(), integ, dest, &entities.NotificationPayload{Title: "T"})
	if err == nil {
		t.Fatal("expected error")
	}
	if result == nil || result.Code != http.StatusUnauthorized {
		t.Errorf("Code = %v, want 401", result)
	}
}

func TestListTargets_ReturnsEmpty(t *testing.T) {
	c := newTestClient(t)
	targets, err := c.ListTargets(t.Context(), &entities.Integration{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 0 {
		t.Errorf("expected empty targets, got %d", len(targets))
	}
}
