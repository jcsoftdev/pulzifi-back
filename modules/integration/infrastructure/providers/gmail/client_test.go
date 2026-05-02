package gmailprovider

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
	if got := q.Get("client_id"); got != "test-client-id" {
		t.Errorf("client_id = %q, want %q", got, "test-client-id")
	}
	if !strings.Contains(q.Get("scope"), "gmail.send") {
		t.Errorf("scope %q should contain gmail.send", q.Get("scope"))
	}
	if q.Get("access_type") != "offline" {
		t.Errorf("access_type = %q, want offline", q.Get("access_type"))
	}
	if q.Get("prompt") != "consent" {
		t.Errorf("prompt = %q, want consent", q.Get("prompt"))
	}
}

func TestHandleCallback_Success(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			// Token exchange
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"access_token":"tok-abc","refresh_token":"ref-xyz","expires_in":3600}`))
		} else {
			// Userinfo
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"email":"user@gmail.com"}`))
		}
	}))
	defer srv.Close()

	origToken, origUserinfo := TokenURL, UserinfoURL
	TokenURL, UserinfoURL = srv.URL, srv.URL
	t.Cleanup(func() { TokenURL, UserinfoURL = origToken, origUserinfo })

	c := newTestClient(t)
	result, err := c.HandleCallback(t.Context(), "code-abc", "https://example.com/callback")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "tok-abc" {
		t.Errorf("AccessToken = %q, want tok-abc", result.AccessToken)
	}
	if result.RefreshToken != "ref-xyz" {
		t.Errorf("RefreshToken = %q, want ref-xyz", result.RefreshToken)
	}
	if result.ExpiresAt == nil {
		t.Error("ExpiresAt should be non-nil")
	}
	if email, _ := result.ProviderMeta["email"].(string); email != "user@gmail.com" {
		t.Errorf("provider_meta.email = %q, want user@gmail.com", email)
	}
}

func TestHandleCallback_TokenError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer srv.Close()

	origToken := TokenURL
	TokenURL = srv.URL
	t.Cleanup(func() { TokenURL = origToken })

	c := newTestClient(t)
	_, err := c.HandleCallback(t.Context(), "bad-code", "https://example.com/callback")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid_grant") {
		t.Errorf("error %q should contain invalid_grant", err.Error())
	}
}

func TestHandleCallback_UserinfoError(t *testing.T) {
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"tok","refresh_token":"ref","expires_in":3600}`))
	}))
	defer tokenSrv.Close()

	userSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer userSrv.Close()

	origToken, origUserinfo := TokenURL, UserinfoURL
	TokenURL, UserinfoURL = tokenSrv.URL, userSrv.URL
	t.Cleanup(func() { TokenURL, UserinfoURL = origToken, origUserinfo })

	c := newTestClient(t)
	_, err := c.HandleCallback(t.Context(), "code", "https://example.com/callback")
	if err == nil {
		t.Fatal("expected error when userinfo fails, got nil")
	}
}

func TestRefreshAccessToken_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		vals, _ := url.ParseQuery(string(body))
		if vals.Get("grant_type") != "refresh_token" {
			t.Errorf("grant_type = %q, want refresh_token", vals.Get("grant_type"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-tok","expires_in":3600}`))
	}))
	defer srv.Close()

	orig := TokenURL
	TokenURL = srv.URL
	t.Cleanup(func() { TokenURL = orig })

	c := newTestClient(t)
	result, err := c.RefreshAccessToken(t.Context(), "ref-tok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "new-tok" {
		t.Errorf("AccessToken = %q, want new-tok", result.AccessToken)
	}
}

func TestSend_Success(t *testing.T) {
	var capturedBody map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-tok" {
			t.Errorf("Authorization = %q, want Bearer test-tok", auth)
		}
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		// Gmail send returns 200 with message ID
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"msg-1"}`))
	}))
	defer srv.Close()

	orig := SendURL
	SendURL = srv.URL
	t.Cleanup(func() { SendURL = orig })

	c := newTestClient(t)
	integ := &entities.Integration{
		ID:           uuid.New(),
		AccessToken:  "test-tok",
		ProviderMeta: map[string]any{"email": "sender@gmail.com"},
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
	// raw field must be non-empty (base64url encoded RFC 2822 message)
	if capturedBody["raw"] == "" {
		t.Error("raw field should be non-empty")
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

func TestSend_NoRecipients_ReturnsError(t *testing.T) {
	c := newTestClient(t)
	integ := &entities.Integration{
		AccessToken:  "tok",
		ProviderMeta: map[string]any{"email": "sender@gmail.com"},
	}
	dest := &entities.Destination{Target: map[string]any{"emails": []any{}}}
	_, err := c.Send(t.Context(), integ, dest, &entities.NotificationPayload{})
	if err == nil {
		t.Fatal("expected error for empty recipients")
	}
}

func TestSend_Unauthorized_ReturnsCode401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_credentials"}`))
	}))
	defer srv.Close()

	orig := SendURL
	SendURL = srv.URL
	t.Cleanup(func() { SendURL = orig })

	c := newTestClient(t)
	integ := &entities.Integration{
		AccessToken:  "bad-tok",
		ProviderMeta: map[string]any{"email": "sender@gmail.com"},
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
