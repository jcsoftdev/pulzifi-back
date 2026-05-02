package discordprovider_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	discordprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/discord"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
)

func newClient() *discordprovider.Client {
	return discordprovider.New("test-client-id", "test-client-secret")
}

// TestDiscord_OAuthAuthorizeURL_IncludesState verifies the authorization URL
// contains all required OAuth query parameters.
func TestDiscord_OAuthAuthorizeURL_IncludesState(t *testing.T) {
	c := newClient()
	rawURL, err := c.OAuthAuthorizeURL("st4te", "https://example.com/cb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("invalid URL: %v", err)
	}

	q := parsed.Query()
	if got := q.Get("client_id"); got != "test-client-id" {
		t.Errorf("client_id: want test-client-id, got %q", got)
	}
	if got := q.Get("scope"); got != discordprovider.Scopes {
		t.Errorf("scope: want %q, got %q", discordprovider.Scopes, got)
	}
	if got := q.Get("response_type"); got != "code" {
		t.Errorf("response_type: want code, got %q", got)
	}
	if got := q.Get("redirect_uri"); got != "https://example.com/cb" {
		t.Errorf("redirect_uri: want https://example.com/cb, got %q", got)
	}
	if got := q.Get("state"); got != "st4te" {
		t.Errorf("state: want st4te, got %q", got)
	}
}

// TestDiscord_HandleCallback_Success verifies successful OAuth callback with webhook in response.
func TestDiscord_HandleCallback_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
			t.Errorf("Content-Type: want application/x-www-form-urlencoded, got %q", ct)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		if got := r.FormValue("client_id"); got != "test-client-id" {
			t.Errorf("form client_id: want test-client-id, got %q", got)
		}
		if got := r.FormValue("grant_type"); got != "authorization_code" {
			t.Errorf("form grant_type: want authorization_code, got %q", got)
		}
		if got := r.FormValue("code"); got != "mycode" {
			t.Errorf("form code: want mycode, got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{
			"access_token": "a-token",
			"scope": "webhook.incoming",
			"webhook": {
				"id": "123",
				"url": "https://discord.com/api/webhooks/123/abc",
				"channel_id": "c1",
				"name": "general",
				"guild_id": "g1"
			}
		}`)
	}))
	defer srv.Close()

	orig := discordprovider.TokenURL
	discordprovider.TokenURL = srv.URL
	t.Cleanup(func() { discordprovider.TokenURL = orig })

	c := newClient()
	result, err := c.HandleCallback(context.Background(), "mycode", "https://example.com/cb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "a-token" {
		t.Errorf("AccessToken: want a-token, got %q", result.AccessToken)
	}
	if wURL, _ := result.ProviderMeta["webhook_url"].(string); wURL != "https://discord.com/api/webhooks/123/abc" {
		t.Errorf("webhook_url: want https://discord.com/api/webhooks/123/abc, got %q", wURL)
	}
	if chID, _ := result.ProviderMeta["channel_id"].(string); chID != "c1" {
		t.Errorf("channel_id: want c1, got %q", chID)
	}
}

// TestDiscord_HandleCallback_Failure verifies an OAuth error response is surfaced.
func TestDiscord_HandleCallback_Failure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(w, `{"error":"invalid_grant"}`)
	}))
	defer srv.Close()

	orig := discordprovider.TokenURL
	discordprovider.TokenURL = srv.URL
	t.Cleanup(func() { discordprovider.TokenURL = orig })

	c := newClient()
	_, err := c.HandleCallback(context.Background(), "badcode", "https://example.com/cb")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid_grant") {
		t.Errorf("error should contain invalid_grant, got: %v", err)
	}
}

// TestDiscord_HandleCallback_NoWebhook verifies an error when no webhook is returned.
func TestDiscord_HandleCallback_NoWebhook(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"access_token":"x","scope":"webhook.incoming"}`)
	}))
	defer srv.Close()

	orig := discordprovider.TokenURL
	discordprovider.TokenURL = srv.URL
	t.Cleanup(func() { discordprovider.TokenURL = orig })

	c := newClient()
	_, err := c.HandleCallback(context.Background(), "code", "https://example.com/cb")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no webhook") {
		t.Errorf("error should contain 'no webhook', got: %v", err)
	}
}

// TestDiscord_Send_Success verifies a successful POST to the Discord webhook endpoint.
func TestDiscord_Send_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type: want application/json, got %q", ct)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		embeds, ok := body["embeds"].([]any)
		if !ok || len(embeds) == 0 {
			t.Fatalf("expected embeds array with at least one element")
		}
		embed, _ := embeds[0].(map[string]any)
		if embed["title"] == nil {
			t.Error("embed missing title")
		}
		if embed["description"] == nil {
			t.Error("embed missing description")
		}
		if embed["url"] == nil {
			t.Error("embed missing url")
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newClient()
	dest := &entities.Destination{
		Target: map[string]any{"webhook_url": srv.URL},
	}
	payload := &entities.NotificationPayload{
		Title:    "Alert fired",
		Body:     "Something changed",
		PageURL:  "https://example.com/page",
		Severity: "info",
	}

	result, err := c.Send(context.Background(), nil, dest, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Code != http.StatusNoContent {
		t.Errorf("Code: want 204, got %d", result.Code)
	}
}

// TestDiscord_Send_AuthFail401 verifies 401 responses are handled and the webhook URL is redacted.
func TestDiscord_Send_AuthFail401(t *testing.T) {
	var capturedWebhookURL string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedWebhookURL = "http://" + r.Host + r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		// Include the webhook URL in the response body to test redaction
		_, _ = io.WriteString(w, `{"message":"401: Unauthorized","code":0,"webhook_url":"`+capturedWebhookURL+`"}`)
	}))
	defer srv.Close()

	webhookURL := srv.URL
	c := newClient()
	dest := &entities.Destination{
		Target: map[string]any{"webhook_url": webhookURL},
	}
	payload := &entities.NotificationPayload{
		Title:    "Alert",
		Body:     "Body",
		Severity: "critical",
	}

	result, err := c.Send(context.Background(), nil, dest, payload)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result.Code != http.StatusUnauthorized {
		t.Errorf("Code: want 401, got %d", result.Code)
	}
	if strings.Contains(result.BodySnip, webhookURL) {
		t.Errorf("BodySnip must not contain the raw webhook URL; got: %s", result.BodySnip)
	}
	if !strings.Contains(result.BodySnip, "***WEBHOOK_URL_REDACTED***") {
		t.Errorf("BodySnip should contain redaction placeholder; got: %s", result.BodySnip)
	}
}

// TestDiscord_Send_WebhookGone404 verifies 404 responses return an error.
func TestDiscord_Send_WebhookGone404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newClient()
	dest := &entities.Destination{
		Target: map[string]any{"webhook_url": srv.URL},
	}
	payload := &entities.NotificationPayload{Title: "Test", Body: "Body", Severity: "info"}

	result, err := c.Send(context.Background(), nil, dest, payload)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result.Code != http.StatusNotFound {
		t.Errorf("Code: want 404, got %d", result.Code)
	}
}

// TestDiscord_Send_FallbackToIntegration verifies Send falls back to integ.ProviderMeta
// when dest.Target has no webhook_url.
func TestDiscord_Send_FallbackToIntegration(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	c := newClient()
	integ := &entities.Integration{
		ProviderMeta: map[string]any{"webhook_url": srv.URL},
	}
	dest := &entities.Destination{
		Target: nil,
	}
	payload := &entities.NotificationPayload{Title: "Fallback test", Body: "Body", Severity: "info"}

	result, err := c.Send(context.Background(), integ, dest, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Code != http.StatusNoContent {
		t.Errorf("Code: want 204, got %d", result.Code)
	}
}

// TestDiscord_Send_NoWebhookURL verifies Send returns an error when no webhook URL is available.
func TestDiscord_Send_NoWebhookURL(t *testing.T) {
	c := newClient()
	dest := &entities.Destination{Target: nil}
	payload := &entities.NotificationPayload{Title: "Test", Body: "Body", Severity: "info"}

	_, err := c.Send(context.Background(), nil, dest, payload)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "discord: no webhook url") {
		t.Errorf("error should contain 'discord: no webhook url', got: %v", err)
	}
}

// TestDiscord_ListTargets_SingleWebhook verifies a single target is returned from ProviderMeta.
func TestDiscord_ListTargets_SingleWebhook(t *testing.T) {
	c := newClient()
	integ := &entities.Integration{
		ProviderMeta: map[string]any{
			"webhook_url":  "https://discord.com/api/webhooks/999/xyz",
			"channel_name": "alerts",
		},
	}

	targets, err := c.ListTargets(context.Background(), integ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 1 {
		t.Fatalf("expected 1 target, got %d", len(targets))
	}
	if targets[0].Name != "#alerts" {
		t.Errorf("Name: want #alerts, got %q", targets[0].Name)
	}
	if wURL, _ := targets[0].Meta["webhook_url"].(string); wURL != "https://discord.com/api/webhooks/999/xyz" {
		t.Errorf("Meta webhook_url: want https://discord.com/api/webhooks/999/xyz, got %q", wURL)
	}
}

// TestDiscord_ListTargets_NoIntegration verifies nil integration returns nil, nil.
func TestDiscord_ListTargets_NoIntegration(t *testing.T) {
	c := newClient()
	targets, err := c.ListTargets(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if targets != nil {
		t.Errorf("expected nil targets, got %v", targets)
	}
}
