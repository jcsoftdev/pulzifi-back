package slackprovider

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

// newTestClient returns a Client wired to an httptest.Server for testing.
// The caller is responsible for closing the server.
func newTestClient(t *testing.T) *Client {
	t.Helper()
	return New("test-client-id", "test-client-secret")
}

func TestOAuthAuthorizeURL_IncludesState(t *testing.T) {
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
	if got := q.Get("scope"); got != Scopes {
		t.Errorf("scope = %q, want %q", got, Scopes)
	}
	if got := q.Get("redirect_uri"); got != "https://example.com/callback" {
		t.Errorf("redirect_uri = %q, want %q", got, "https://example.com/callback")
	}
}

func TestHandleCallback_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/x-www-form-urlencoded") {
			t.Errorf("Content-Type = %q, want form-urlencoded", ct)
		}
		body, _ := io.ReadAll(r.Body)
		vals, _ := url.ParseQuery(string(body))
		if vals.Get("code") == "" {
			t.Error("missing code in request body")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"access_token":"xoxb-X","team":{"id":"T1","name":"My Team"},"bot_user_id":"B1"}`))
	}))
	defer srv.Close()

	orig := TokenURL
	TokenURL = srv.URL
	t.Cleanup(func() { TokenURL = orig })

	c := newTestClient(t)
	result, err := c.HandleCallback(t.Context(), "code-abc", "https://example.com/callback")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "xoxb-X" {
		t.Errorf("AccessToken = %q, want %q", result.AccessToken, "xoxb-X")
	}
	if tid, _ := result.ProviderMeta["team_id"].(string); tid != "T1" {
		t.Errorf("team_id = %q, want %q", tid, "T1")
	}
}

func TestHandleCallback_FailureMapsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":false,"error":"invalid_code"}`))
	}))
	defer srv.Close()

	orig := TokenURL
	TokenURL = srv.URL
	t.Cleanup(func() { TokenURL = orig })

	c := newTestClient(t)
	_, err := c.HandleCallback(t.Context(), "bad-code", "https://example.com/callback")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid_code") {
		t.Errorf("error %q should contain %q", err.Error(), "invalid_code")
	}
}

func TestSend_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "Bearer xoxb-test" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer xoxb-test")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if ch, _ := body["channel"].(string); ch != "C123" {
			t.Errorf("channel = %q, want %q", ch, "C123")
		}
		blocks, _ := body["blocks"].([]any)
		if len(blocks) < 2 {
			t.Errorf("blocks length = %d, want >= 2", len(blocks))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	orig := PostMessage
	PostMessage = srv.URL
	t.Cleanup(func() { PostMessage = orig })

	c := newTestClient(t)
	integ := &entities.Integration{
		ID:          uuid.New(),
		AccessToken: "xoxb-test",
	}
	dest := &entities.Destination{
		ID:     uuid.New(),
		Target: map[string]any{"channel_id": "C123"},
	}
	payload := &entities.NotificationPayload{
		Title:   "Test Alert",
		Body:    "Something changed",
		PageURL: "https://example.com/page",
	}

	result, err := c.Send(t.Context(), integ, dest, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Code != http.StatusOK {
		t.Errorf("Code = %d, want %d", result.Code, http.StatusOK)
	}
}

func TestSend_AuthInvalid(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":false,"error":"invalid_auth"}`))
	}))
	defer srv.Close()

	orig := PostMessage
	PostMessage = srv.URL
	t.Cleanup(func() { PostMessage = orig })

	c := newTestClient(t)
	integ := &entities.Integration{
		ID:          uuid.New(),
		AccessToken: "bad-token",
	}
	dest := &entities.Destination{
		ID:     uuid.New(),
		Target: map[string]any{"channel_id": "C123"},
	}
	payload := &entities.NotificationPayload{Title: "Test", Body: "Body"}

	result, err := c.Send(t.Context(), integ, dest, payload)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result == nil {
		t.Fatal("expected non-nil DeliveryResult")
	}
	if result.Code != http.StatusUnauthorized {
		t.Errorf("Code = %d, want %d", result.Code, http.StatusUnauthorized)
	}
	if result.BodySnip == "" {
		t.Error("BodySnip should be non-empty")
	}
	if !strings.Contains(err.Error(), "invalid_auth") {
		t.Errorf("error %q should contain %q", err.Error(), "invalid_auth")
	}
}

func TestListTargets_Pagination(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		cursor := r.URL.Query().Get("cursor")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if cursor == "" {
			// First page: 2 channels + next_cursor
			_, _ = w.Write([]byte(`{
				"ok": true,
				"channels": [
					{"id": "C001", "name": "general"},
					{"id": "C002", "name": "random"}
				],
				"response_metadata": {"next_cursor": "cur2"}
			}`))
		} else {
			// Second page: 1 channel + empty cursor
			_, _ = w.Write([]byte(`{
				"ok": true,
				"channels": [
					{"id": "C003", "name": "ops"}
				],
				"response_metadata": {"next_cursor": ""}
			}`))
		}
	}))
	defer srv.Close()

	orig := ListChannels
	ListChannels = srv.URL
	t.Cleanup(func() { ListChannels = orig })

	c := newTestClient(t)
	integ := &entities.Integration{
		ID:          uuid.New(),
		AccessToken: "xoxb-test",
	}

	targets, err := c.ListTargets(t.Context(), integ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 3 {
		t.Errorf("got %d targets, want 3", len(targets))
	}
	for _, tgt := range targets {
		if !strings.HasPrefix(tgt.Name, "#") {
			t.Errorf("target name %q does not start with '#'", tgt.Name)
		}
	}
	if callCount != 2 {
		t.Errorf("handler called %d times, want 2", callCount)
	}
}
