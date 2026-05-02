package sheetsprovider_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	sheetsprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/sheets"
)

func newTestClient(t *testing.T) *sheetsprovider.Client {
	t.Helper()
	return sheetsprovider.New("test-client-id", "test-client-secret")
}

// TestSheets_OAuthAuthorizeURL_IncludesAccessTypeOffline verifies all required OAuth params.
func TestSheets_OAuthAuthorizeURL_IncludesAccessTypeOffline(t *testing.T) {
	c := newTestClient(t)
	rawURL, err := c.OAuthAuthorizeURL("st4te", "https://r.example/cb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("invalid URL: %v", err)
	}

	q := parsed.Query()
	checks := map[string]string{
		"client_id":              "test-client-id",
		"response_type":          "code",
		"redirect_uri":           "https://r.example/cb",
		"state":                  "st4te",
		"access_type":            "offline",
		"prompt":                 "consent",
		"include_granted_scopes": "true",
	}
	for param, want := range checks {
		if got := q.Get(param); got != want {
			t.Errorf("param %q = %q, want %q", param, got, want)
		}
	}

	scope := q.Get("scope")
	if !strings.Contains(scope, "spreadsheets") {
		t.Errorf("scope %q should contain 'spreadsheets'", scope)
	}
	if !strings.Contains(scope, "drive.readonly") {
		t.Errorf("scope %q should contain 'drive.readonly'", scope)
	}
}

// TestSheets_HandleCallback_Success verifies a successful OAuth callback exchange.
func TestSheets_HandleCallback_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); !strings.Contains(ct, "application/x-www-form-urlencoded") {
			t.Errorf("Content-Type = %q, want form-urlencoded", ct)
		}
		body, _ := io.ReadAll(r.Body)
		vals, _ := url.ParseQuery(string(body))
		if vals.Get("code") == "" {
			t.Error("missing code in request body")
		}
		if vals.Get("grant_type") != "authorization_code" {
			t.Errorf("grant_type = %q, want authorization_code", vals.Get("grant_type"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"a","refresh_token":"r","expires_in":3600,"scope":"...","token_type":"Bearer"}`))
	}))
	defer srv.Close()

	orig := sheetsprovider.TokenURL
	sheetsprovider.TokenURL = srv.URL
	t.Cleanup(func() { sheetsprovider.TokenURL = orig })

	c := newTestClient(t)
	result, err := c.HandleCallback(t.Context(), "code-abc", "https://r.example/cb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "a" {
		t.Errorf("AccessToken = %q, want %q", result.AccessToken, "a")
	}
	if result.RefreshToken != "r" {
		t.Errorf("RefreshToken = %q, want %q", result.RefreshToken, "r")
	}
	if result.ExpiresAt == nil {
		t.Fatal("ExpiresAt should not be nil")
	}
	// ExpiresAt should be approximately now+3600s (within 5s tolerance)
	expected := time.Now().UTC().Add(3600 * time.Second)
	diff := result.ExpiresAt.Sub(expected)
	if diff < -5*time.Second || diff > 5*time.Second {
		t.Errorf("ExpiresAt diff = %v, want within 5s of +3600s", diff)
	}
	if result.ProviderMeta == nil {
		t.Error("ProviderMeta should not be nil")
	}
}

// TestSheets_HandleCallback_NoRefreshToken_Error verifies error when refresh_token is absent.
func TestSheets_HandleCallback_NoRefreshToken_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"a","expires_in":3600,"token_type":"Bearer"}`))
	}))
	defer srv.Close()

	orig := sheetsprovider.TokenURL
	sheetsprovider.TokenURL = srv.URL
	t.Cleanup(func() { sheetsprovider.TokenURL = orig })

	c := newTestClient(t)
	_, err := c.HandleCallback(t.Context(), "code-abc", "https://r.example/cb")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no refresh_token") {
		t.Errorf("error %q should contain 'no refresh_token'", err.Error())
	}
}

// TestSheets_HandleCallback_ErrorResponse verifies HTTP error status or error field is propagated.
func TestSheets_HandleCallback_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer srv.Close()

	orig := sheetsprovider.TokenURL
	sheetsprovider.TokenURL = srv.URL
	t.Cleanup(func() { sheetsprovider.TokenURL = orig })

	c := newTestClient(t)
	_, err := c.HandleCallback(t.Context(), "bad-code", "https://r.example/cb")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "400") && !strings.Contains(err.Error(), "invalid_grant") {
		t.Errorf("error %q should contain '400' or 'invalid_grant'", err.Error())
	}
}

// TestSheets_RefreshAccessToken_Success verifies a successful token refresh.
func TestSheets_RefreshAccessToken_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		vals, _ := url.ParseQuery(string(body))
		if vals.Get("grant_type") != "refresh_token" {
			t.Errorf("grant_type = %q, want refresh_token", vals.Get("grant_type"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"new-access","expires_in":3600,"token_type":"Bearer"}`))
	}))
	defer srv.Close()

	orig := sheetsprovider.TokenURL
	sheetsprovider.TokenURL = srv.URL
	t.Cleanup(func() { sheetsprovider.TokenURL = orig })

	c := newTestClient(t)
	before := time.Now().UTC()
	result, err := c.RefreshAccessToken(t.Context(), "my-refresh-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "new-access" {
		t.Errorf("AccessToken = %q, want %q", result.AccessToken, "new-access")
	}
	if result.ExpiresAt == nil || !result.ExpiresAt.After(before) {
		t.Error("ExpiresAt should be in the future")
	}
}

// TestSheets_RefreshAccessToken_RotatedRefresh verifies a rotated refresh token is returned.
func TestSheets_RefreshAccessToken_RotatedRefresh(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"new-access","refresh_token":"new-refresh","expires_in":3600,"token_type":"Bearer"}`))
	}))
	defer srv.Close()

	orig := sheetsprovider.TokenURL
	sheetsprovider.TokenURL = srv.URL
	t.Cleanup(func() { sheetsprovider.TokenURL = orig })

	c := newTestClient(t)
	result, err := c.RefreshAccessToken(t.Context(), "old-refresh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RefreshToken != "new-refresh" {
		t.Errorf("RefreshToken = %q, want %q", result.RefreshToken, "new-refresh")
	}
}

// TestSheets_RefreshAccessToken_NoNewRefresh_PreservesEmpty verifies no error when response omits refresh_token.
func TestSheets_RefreshAccessToken_NoNewRefresh_PreservesEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"access_token":"new-access","expires_in":3600,"token_type":"Bearer"}`))
	}))
	defer srv.Close()

	orig := sheetsprovider.TokenURL
	sheetsprovider.TokenURL = srv.URL
	t.Cleanup(func() { sheetsprovider.TokenURL = orig })

	c := newTestClient(t)
	result, err := c.RefreshAccessToken(t.Context(), "my-refresh")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.RefreshToken != "" {
		t.Errorf("RefreshToken = %q, want empty string", result.RefreshToken)
	}
}

// TestSheets_RefreshAccessToken_RevokedToken_Error verifies a revoked token returns an error.
func TestSheets_RefreshAccessToken_RevokedToken_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer srv.Close()

	orig := sheetsprovider.TokenURL
	sheetsprovider.TokenURL = srv.URL
	t.Cleanup(func() { sheetsprovider.TokenURL = orig })

	c := newTestClient(t)
	_, err := c.RefreshAccessToken(t.Context(), "revoked-refresh")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Errorf("error %q should contain '400'", err.Error())
	}
}

// TestSheets_ListTargets_Pagination verifies paginated Drive file listing.
func TestSheets_ListTargets_Pagination(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		pageToken := r.URL.Query().Get("pageToken")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if pageToken == "" {
			// First page: 2 files + nextPageToken
			_, _ = w.Write([]byte(`{
				"files": [
					{"id": "id1", "name": "Sheet One"},
					{"id": "id2", "name": "Sheet Two"}
				],
				"nextPageToken": "page2"
			}`))
		} else {
			// Second page: assert pageToken param and return 1 file
			if pageToken != "page2" {
				t.Errorf("pageToken = %q, want %q", pageToken, "page2")
			}
			_, _ = w.Write([]byte(`{
				"files": [
					{"id": "id3", "name": "Sheet Three"}
				]
			}`))
		}
	}))
	defer srv.Close()

	orig := sheetsprovider.DriveListURL
	sheetsprovider.DriveListURL = srv.URL
	t.Cleanup(func() { sheetsprovider.DriveListURL = orig })

	c := newTestClient(t)
	integ := &entities.Integration{
		ID:          uuid.New(),
		AccessToken: "bearer-token",
	}

	targets, err := c.ListTargets(t.Context(), integ)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(targets) != 3 {
		t.Errorf("got %d targets, want 3", len(targets))
	}
	if callCount != 2 {
		t.Errorf("handler called %d times, want 2", callCount)
	}
	names := []string{"Sheet One", "Sheet Two", "Sheet Three"}
	for i, tgt := range targets {
		if tgt.Name != names[i] {
			t.Errorf("targets[%d].Name = %q, want %q", i, tgt.Name, names[i])
		}
	}
}

// TestSheets_ListTargets_AuthFail401 verifies 401 from Drive returns auth failure error.
func TestSheets_ListTargets_AuthFail401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"code":401,"message":"Request had invalid authentication credentials."}}`))
	}))
	defer srv.Close()

	orig := sheetsprovider.DriveListURL
	sheetsprovider.DriveListURL = srv.URL
	t.Cleanup(func() { sheetsprovider.DriveListURL = orig })

	c := newTestClient(t)
	integ := &entities.Integration{
		ID:          uuid.New(),
		AccessToken: "bad-token",
	}

	_, err := c.ListTargets(t.Context(), integ)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "auth failure") && !strings.Contains(err.Error(), "401") {
		t.Errorf("error %q should contain 'auth failure' or '401'", err.Error())
	}
}

// TestSheets_ListTargets_NoIntegration verifies nil integration returns error.
func TestSheets_ListTargets_NoIntegration(t *testing.T) {
	c := newTestClient(t)
	_, err := c.ListTargets(t.Context(), nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no access token") {
		t.Errorf("error %q should contain 'no access token'", err.Error())
	}
}

// TestSheets_Send_Success verifies a successful append to Google Sheets.
func TestSheets_Send_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify URL contains expected path and query params
		if !strings.Contains(r.URL.Path, "/spreadsheets/abc/values/") {
			t.Errorf("URL path %q missing expected spreadsheet path", r.URL.Path)
		}
		if r.URL.Query().Get("valueInputOption") != "USER_ENTERED" {
			t.Errorf("valueInputOption = %q, want USER_ENTERED", r.URL.Query().Get("valueInputOption"))
		}
		if r.URL.Query().Get("insertDataOption") != "INSERT_ROWS" {
			t.Errorf("insertDataOption = %q, want INSERT_ROWS", r.URL.Query().Get("insertDataOption"))
		}
		// Verify Authorization header
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want 'Bearer test-token'", auth)
		}
		// Decode body and assert values structure
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
		}
		values, _ := body["values"].([]any)
		if len(values) != 1 {
			t.Errorf("values length = %d, want 1", len(values))
		} else {
			row, _ := values[0].([]any)
			if len(row) != 5 {
				t.Errorf("row length = %d, want 5", len(row))
			}
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	orig := sheetsprovider.SheetsBase
	sheetsprovider.SheetsBase = srv.URL
	t.Cleanup(func() { sheetsprovider.SheetsBase = orig })

	c := newTestClient(t)
	integ := &entities.Integration{
		ID:          uuid.New(),
		AccessToken: "test-token",
	}
	dest := &entities.Destination{
		ID: uuid.New(),
		Target: map[string]any{
			"spreadsheet_id": "abc",
			"tab_name":       "Sheet1",
		},
	}
	payload := &entities.NotificationPayload{
		EventType: "change_detected",
		PageURL:   "https://example.com/page",
		Title:     "Test Alert",
		Severity:  "warning",
	}

	result, err := c.Send(t.Context(), integ, dest, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Code != http.StatusOK {
		t.Errorf("Code = %d, want %d", result.Code, http.StatusOK)
	}
}

// TestSheets_Send_AuthFail401 verifies 401 response is mapped to unauthorized result.
func TestSheets_Send_AuthFail401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"code":401,"message":"Invalid Credentials"}}`))
	}))
	defer srv.Close()

	orig := sheetsprovider.SheetsBase
	sheetsprovider.SheetsBase = srv.URL
	t.Cleanup(func() { sheetsprovider.SheetsBase = orig })

	c := newTestClient(t)
	integ := &entities.Integration{ID: uuid.New(), AccessToken: "bad-token"}
	dest := &entities.Destination{
		ID:     uuid.New(),
		Target: map[string]any{"spreadsheet_id": "abc", "tab_name": "Sheet1"},
	}
	payload := &entities.NotificationPayload{Title: "Test", EventType: "test"}

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
	if !strings.Contains(err.Error(), "auth") {
		t.Errorf("error %q should contain 'auth'", err.Error())
	}
}

// TestSheets_Send_SpreadsheetNotFound404 verifies 404 response is propagated correctly.
func TestSheets_Send_SpreadsheetNotFound404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":{"message":"Not Found"}}`))
	}))
	defer srv.Close()

	orig := sheetsprovider.SheetsBase
	sheetsprovider.SheetsBase = srv.URL
	t.Cleanup(func() { sheetsprovider.SheetsBase = orig })

	c := newTestClient(t)
	integ := &entities.Integration{ID: uuid.New(), AccessToken: "valid-token"}
	dest := &entities.Destination{
		ID:     uuid.New(),
		Target: map[string]any{"spreadsheet_id": "nonexistent", "tab_name": "Sheet1"},
	}
	payload := &entities.NotificationPayload{Title: "Test", EventType: "test"}

	result, err := c.Send(t.Context(), integ, dest, payload)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result == nil {
		t.Fatal("expected non-nil DeliveryResult")
	}
	if result.Code != http.StatusNotFound {
		t.Errorf("Code = %d, want %d", result.Code, http.StatusNotFound)
	}
}

// TestSheets_Send_MissingSpreadsheetID verifies error when spreadsheet_id is absent.
func TestSheets_Send_MissingSpreadsheetID(t *testing.T) {
	c := newTestClient(t)
	integ := &entities.Integration{ID: uuid.New(), AccessToken: "valid-token"}
	dest := &entities.Destination{
		ID:     uuid.New(),
		Target: map[string]any{"tab_name": "Sheet1"},
	}
	payload := &entities.NotificationPayload{Title: "Test"}

	result, err := c.Send(t.Context(), integ, dest, payload)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Error("expected nil DeliveryResult when validation fails before HTTP call")
	}
	if !strings.Contains(err.Error(), "missing spreadsheet_id") {
		t.Errorf("error %q should contain 'missing spreadsheet_id'", err.Error())
	}
}

// TestSheets_Send_MissingTabName verifies error when tab_name is absent.
func TestSheets_Send_MissingTabName(t *testing.T) {
	c := newTestClient(t)
	integ := &entities.Integration{ID: uuid.New(), AccessToken: "valid-token"}
	dest := &entities.Destination{
		ID:     uuid.New(),
		Target: map[string]any{"spreadsheet_id": "abc"},
	}
	payload := &entities.NotificationPayload{Title: "Test"}

	result, err := c.Send(t.Context(), integ, dest, payload)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Error("expected nil DeliveryResult when validation fails before HTTP call")
	}
	if !strings.Contains(err.Error(), "missing tab_name") {
		t.Errorf("error %q should contain 'missing tab_name'", err.Error())
	}
}
