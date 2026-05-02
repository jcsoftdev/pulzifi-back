package teamsprovider_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	teamsprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/teams"
)

// helpers

func newClient() *teamsprovider.Client {
	return teamsprovider.New("test-client-id", "test-client-secret")
}

func restoreURLs(origAuth, origToken, origGraph, origConsent string) func() {
	return func() {
		teamsprovider.AuthorizeURL = origAuth
		teamsprovider.TokenURL = origToken
		teamsprovider.GraphBase = origGraph
		teamsprovider.AdminConsentURLBase = origConsent
	}
}

// 1. TestTeams_OAuthAuthorizeURL_IncludesOfflineAccess
func TestTeams_OAuthAuthorizeURL_IncludesOfflineAccess(t *testing.T) {
	c := newClient()
	u, err := c.OAuthAuthorizeURL("mystate", "https://app.example.com/callback")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(u, "offline_access") {
		t.Errorf("authorize URL missing offline_access scope: %s", u)
	}
	if !strings.Contains(u, "prompt=select_account") {
		t.Errorf("authorize URL missing prompt=select_account: %s", u)
	}
	if !strings.Contains(u, "response_mode=query") {
		t.Errorf("authorize URL missing response_mode=query: %s", u)
	}
}

// 2. TestTeams_HandleCallback_Success
func TestTeams_HandleCallback_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "acc-token-123",
			"refresh_token": "ref-token-456",
			"expires_in":    3600,
			"scope":         teamsprovider.Scopes,
			"token_type":    "Bearer",
		})
	}))
	defer srv.Close()

	orig := teamsprovider.TokenURL
	teamsprovider.TokenURL = srv.URL
	t.Cleanup(func() { teamsprovider.TokenURL = orig })

	c := newClient()
	result, err := c.HandleCallback(context.Background(), "auth-code", "https://redirect.example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "acc-token-123" {
		t.Errorf("expected access_token=acc-token-123, got %q", result.AccessToken)
	}
	if result.RefreshToken != "ref-token-456" {
		t.Errorf("expected refresh_token=ref-token-456, got %q", result.RefreshToken)
	}
	if result.ExpiresAt == nil {
		t.Error("ExpiresAt should not be nil")
	}
}

// 3. TestTeams_HandleCallback_TokenError
func TestTeams_HandleCallback_TokenError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer srv.Close()

	orig := teamsprovider.TokenURL
	teamsprovider.TokenURL = srv.URL
	t.Cleanup(func() { teamsprovider.TokenURL = orig })

	c := newClient()
	_, err := c.HandleCallback(context.Background(), "bad-code", "https://redirect.example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "status=400") {
		t.Errorf("expected error to contain 'status=400', got: %v", err)
	}
}

// 4. TestTeams_BuildAdminConsentURL
func TestTeams_BuildAdminConsentURL(t *testing.T) {
	c := newClient()
	consentErr := c.BuildAdminConsentURL("https://app.example.com/consent-callback")
	if consentErr == nil {
		t.Fatal("expected non-nil ErrConsentRequired")
	}
	u := consentErr.AdminConsentURL
	if !strings.Contains(u, "client_id=test-client-id") {
		t.Errorf("consent URL missing client_id: %s", u)
	}
	if !strings.Contains(u, "redirect_uri=") {
		t.Errorf("consent URL missing redirect_uri: %s", u)
	}
}

// 5. TestTeams_RefreshAccessToken_Success_RotatesRefreshToken
func TestTeams_RefreshAccessToken_Success_RotatesRefreshToken(t *testing.T) {
	origRefresh := "old-refresh-token"
	newRefresh := "new-refresh-token-rotated"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "new-access-token",
			"refresh_token": newRefresh,
			"expires_in":    3600,
			"token_type":    "Bearer",
		})
	}))
	defer srv.Close()

	orig := teamsprovider.TokenURL
	teamsprovider.TokenURL = srv.URL
	t.Cleanup(func() { teamsprovider.TokenURL = orig })

	c := newClient()
	result, err := c.RefreshAccessToken(context.Background(), origRefresh)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RefreshToken == origRefresh {
		t.Errorf("refresh_token was not rotated; still %q", result.RefreshToken)
	}
	if result.RefreshToken != newRefresh {
		t.Errorf("expected new refresh_token=%q, got %q", newRefresh, result.RefreshToken)
	}
}

// 6. TestTeams_RefreshAccessToken_RevokedToken_Error
func TestTeams_RefreshAccessToken_RevokedToken_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_grant","error_description":"Token has been revoked"}`))
	}))
	defer srv.Close()

	orig := teamsprovider.TokenURL
	teamsprovider.TokenURL = srv.URL
	t.Cleanup(func() { teamsprovider.TokenURL = orig })

	c := newClient()
	_, err := c.RefreshAccessToken(context.Background(), "revoked-token")
	if err == nil {
		t.Fatal("expected error for revoked token, got nil")
	}
}

// 7. TestTeams_RefreshAccessToken_EmptyRefresh_Error
func TestTeams_RefreshAccessToken_EmptyRefresh_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// access_token present but no refresh_token
		json.NewEncoder(w).Encode(map[string]any{
			"access_token": "some-access-token",
			"expires_in":   3600,
			"token_type":   "Bearer",
		})
	}))
	defer srv.Close()

	orig := teamsprovider.TokenURL
	teamsprovider.TokenURL = srv.URL
	t.Cleanup(func() { teamsprovider.TokenURL = orig })

	c := newClient()
	_, err := c.RefreshAccessToken(context.Background(), "some-refresh")
	if err == nil {
		t.Fatal("expected error for empty refresh_token, got nil")
	}
	if !strings.Contains(err.Error(), "empty refresh_token") {
		t.Errorf("expected error to contain 'empty refresh_token', got: %v", err)
	}
}

// 8. TestTeams_ListTargets_TwoTeamsThreeChannels
func TestTeams_ListTargets_TwoTeamsThreeChannels(t *testing.T) {
	origGraphBase := teamsprovider.GraphBase

	mux := http.NewServeMux()

	mux.HandleFunc("/me/joinedTeams", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"value": []map[string]any{
				{"id": "teamA", "displayName": "Alpha"},
				{"id": "teamB", "displayName": "Beta"},
			},
		})
	})
	mux.HandleFunc("/teams/teamA/channels", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"value": []map[string]any{
				{"id": "chan-A1", "displayName": "general"},
				{"id": "chan-A2", "displayName": "random"},
			},
		})
	})
	mux.HandleFunc("/teams/teamB/channels", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"value": []map[string]any{
				{"id": "chan-B1", "displayName": "general"},
			},
		})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()
	teamsprovider.GraphBase = srv.URL
	t.Cleanup(func() { teamsprovider.GraphBase = origGraphBase })

	c := newClient()
	integ := &entities.Integration{AccessToken: "tok"}
	targets, err := c.ListTargets(context.Background(), integ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}

	// Build a map for easier assertion
	byID := make(map[string]entities.Target)
	for _, tgt := range targets {
		byID[tgt.ID] = tgt
	}

	if _, ok := byID["teamA|chan-A1"]; !ok {
		t.Errorf("missing target teamA|chan-A1; got IDs: %v", targetIDs(targets))
	}
	if _, ok := byID["teamA|chan-A2"]; !ok {
		t.Errorf("missing target teamA|chan-A2; got IDs: %v", targetIDs(targets))
	}
	if _, ok := byID["teamB|chan-B1"]; !ok {
		t.Errorf("missing target teamB|chan-B1; got IDs: %v", targetIDs(targets))
	}
	if byID["teamA|chan-A1"].Name != "Alpha / #general" {
		t.Errorf("unexpected name: %q", byID["teamA|chan-A1"].Name)
	}
}

func targetIDs(targets []entities.Target) []string {
	ids := make([]string, len(targets))
	for i, t := range targets {
		ids[i] = t.ID
	}
	return ids
}

// 9. TestTeams_ListTargets_Pagination
func TestTeams_ListTargets_Pagination(t *testing.T) {
	origGraphBase := teamsprovider.GraphBase

	var srv *httptest.Server
	mux := http.NewServeMux()

	// First page of joinedTeams — returns nextLink pointing at second page
	mux.HandleFunc("/me/joinedTeams", func(w http.ResponseWriter, r *http.Request) {
		nextLink := srv.URL + "/me/joinedTeams/page2"
		json.NewEncoder(w).Encode(map[string]any{
			"value": []map[string]any{
				{"id": "teamA", "displayName": "Alpha"},
			},
			"@odata.nextLink": nextLink,
		})
	})
	// Second page (the nextLink URL)
	mux.HandleFunc("/me/joinedTeams/page2", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"value": []map[string]any{
				{"id": "teamB", "displayName": "Beta"},
			},
		})
	})
	mux.HandleFunc("/teams/teamA/channels", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"value": []map[string]any{
				{"id": "chan-A1", "displayName": "general"},
			},
		})
	})
	mux.HandleFunc("/teams/teamB/channels", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"value": []map[string]any{
				{"id": "chan-B1", "displayName": "announcements"},
			},
		})
	})

	srv = httptest.NewServer(mux)
	defer srv.Close()
	teamsprovider.GraphBase = srv.URL
	t.Cleanup(func() { teamsprovider.GraphBase = origGraphBase })

	c := newClient()
	integ := &entities.Integration{AccessToken: "tok"}
	targets, err := c.ListTargets(context.Background(), integ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("expected 2 targets (one per team), got %d: %v", len(targets), targetIDs(targets))
	}

	byID := make(map[string]entities.Target)
	for _, tgt := range targets {
		byID[tgt.ID] = tgt
	}
	if _, ok := byID["teamA|chan-A1"]; !ok {
		t.Errorf("missing teamA|chan-A1")
	}
	if _, ok := byID["teamB|chan-B1"]; !ok {
		t.Errorf("missing teamB|chan-B1")
	}
}

// 10. TestTeams_ListTargets_NoIntegration
func TestTeams_ListTargets_NoIntegration(t *testing.T) {
	c := newClient()
	_, err := c.ListTargets(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil integration, got nil")
	}
	if !strings.Contains(err.Error(), "no access token") {
		t.Errorf("expected 'no access token' in error, got: %v", err)
	}
}

// 11. TestTeams_Send_FromCompositeTarget
func TestTeams_Send_FromCompositeTarget(t *testing.T) {
	origGraphBase := teamsprovider.GraphBase

	var capturedPath string
	var capturedBody map[string]any

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&capturedBody)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"msg-1"}`))
	}))
	defer srv.Close()
	teamsprovider.GraphBase = srv.URL
	t.Cleanup(func() { teamsprovider.GraphBase = origGraphBase })

	c := newClient()
	integ := &entities.Integration{AccessToken: "bearer-token"}
	dest := &entities.Destination{
		Target: map[string]any{
			"channel_target": "teamA|chan1",
		},
	}
	payload := &entities.NotificationPayload{
		Title:   "Alert: <Price Change>",
		Body:    "Something happened",
		PageURL: "https://example.com/page",
	}

	result, err := c.Send(context.Background(), integ, dest, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Code != http.StatusCreated {
		t.Errorf("expected code 201, got %d", result.Code)
	}
	if capturedPath != "/teams/teamA/channels/chan1/messages" {
		t.Errorf("unexpected POST path: %s", capturedPath)
	}
	if capturedBody["subject"] != "Alert: <Price Change>" {
		t.Errorf("unexpected subject: %v", capturedBody["subject"])
	}
	bodyObj, ok := capturedBody["body"].(map[string]any)
	if !ok {
		t.Fatalf("body field missing or wrong type")
	}
	if bodyObj["contentType"] != "html" {
		t.Errorf("expected contentType=html, got %v", bodyObj["contentType"])
	}
	content, _ := bodyObj["content"].(string)
	if !strings.Contains(content, "Alert: &lt;Price Change&gt;") {
		t.Errorf("title not HTML-escaped in content: %q", content)
	}
}

// 12. TestTeams_Send_FromSplitTarget
func TestTeams_Send_FromSplitTarget(t *testing.T) {
	origGraphBase := teamsprovider.GraphBase

	var capturedPath string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"msg-2"}`))
	}))
	defer srv.Close()
	teamsprovider.GraphBase = srv.URL
	t.Cleanup(func() { teamsprovider.GraphBase = origGraphBase })

	c := newClient()
	integ := &entities.Integration{AccessToken: "bearer-token"}
	dest := &entities.Destination{
		Target: map[string]any{
			"team_id":    "teamA",
			"channel_id": "chan1",
		},
	}
	payload := &entities.NotificationPayload{
		Title: "Test",
		Body:  "Body text",
	}

	result, err := c.Send(context.Background(), integ, dest, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Code != http.StatusCreated {
		t.Errorf("expected code 201, got %d", result.Code)
	}
	if capturedPath != "/teams/teamA/channels/chan1/messages" {
		t.Errorf("unexpected POST path: %s", capturedPath)
	}
}

// 13. TestTeams_Send_AuthFail401
func TestTeams_Send_AuthFail401(t *testing.T) {
	origGraphBase := teamsprovider.GraphBase

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":{"code":"InvalidAuthenticationToken"}}`))
	}))
	defer srv.Close()
	teamsprovider.GraphBase = srv.URL
	t.Cleanup(func() { teamsprovider.GraphBase = origGraphBase })

	c := newClient()
	integ := &entities.Integration{AccessToken: "expired-token"}
	dest := &entities.Destination{
		Target: map[string]any{"team_id": "teamA", "channel_id": "chan1"},
	}
	payload := &entities.NotificationPayload{Title: "Test", Body: "Body"}

	result, err := c.Send(context.Background(), integ, dest, payload)
	if err == nil {
		t.Fatal("expected error for 401, got nil")
	}
	if result == nil {
		t.Fatal("expected non-nil result even on error")
	}
	if result.Code != 401 {
		t.Errorf("expected result.Code=401, got %d", result.Code)
	}
}

// 14. TestTeams_Send_ChannelNotFound404
func TestTeams_Send_ChannelNotFound404(t *testing.T) {
	origGraphBase := teamsprovider.GraphBase

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":{"code":"ItemNotFound"}}`))
	}))
	defer srv.Close()
	teamsprovider.GraphBase = srv.URL
	t.Cleanup(func() { teamsprovider.GraphBase = origGraphBase })

	c := newClient()
	integ := &entities.Integration{AccessToken: "token"}
	dest := &entities.Destination{
		Target: map[string]any{"team_id": "teamA", "channel_id": "nonexistent"},
	}
	payload := &entities.NotificationPayload{Title: "Test", Body: "Body"}

	result, err := c.Send(context.Background(), integ, dest, payload)
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
	if result == nil {
		t.Fatal("expected non-nil result even on error")
	}
	if result.Code != 404 {
		t.Errorf("expected result.Code=404, got %d", result.Code)
	}
	if !strings.Contains(err.Error(), "status=404") {
		t.Errorf("expected 'status=404' in error, got: %v", err)
	}
}

// 15. TestTeams_Send_MissingTarget
func TestTeams_Send_MissingTarget(t *testing.T) {
	c := newClient()
	integ := &entities.Integration{AccessToken: "token"}
	dest := &entities.Destination{
		Target: map[string]any{},
	}
	payload := &entities.NotificationPayload{Title: "Test", Body: "Body"}

	_, err := c.Send(context.Background(), integ, dest, payload)
	if err == nil {
		t.Fatal("expected error for empty target, got nil")
	}
	if !strings.Contains(err.Error(), "missing team_id/channel_id") {
		t.Errorf("expected 'missing team_id/channel_id' in error, got: %v", err)
	}
}

// Ensure fmt is used (compilation guard).
var _ = fmt.Sprintf
