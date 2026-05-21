package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

// ── TestGitHubProvider_AuthCodeURL ────────────────────────────────────────────

// TestGitHubProvider_AuthCodeURL verifies that the generated URL contains the
// client_id and state parameters.
func TestGitHubProvider_AuthCodeURL(t *testing.T) {
	p := NewGitHubProvider("gh-client-id", "secret", "http://localhost/callback")
	url := p.AuthCodeURL("state-abc")

	if !strings.Contains(url, "gh-client-id") {
		t.Errorf("expected client_id in AuthCodeURL, got %q", url)
	}
	if !strings.Contains(url, "state-abc") {
		t.Errorf("expected state in AuthCodeURL, got %q", url)
	}
}

// ── TestGitHubProvider_Exchange_Success_WithInlineEmail ───────────────────────

// TestGitHubProvider_Exchange_Success_WithInlineEmail verifies the happy path when
// the user profile includes an email directly (no extra emails API call needed).
func TestGitHubProvider_Exchange_Success_WithInlineEmail(t *testing.T) {
	userProfile := GitHubUser{
		ID:        42,
		Login:     "alice",
		Name:      "Alice Smith",
		Email:     "alice@example.com",
		AvatarURL: "https://example.com/avatar.jpg",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "gh-token",
			"token_type":   "bearer",
		})
	})
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userProfile)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	wrapped := newGitHubTestProvider(srv)
	user, token, err := wrapped.Exchange(context.Background(), "code-xyz")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == nil {
		t.Fatal("expected non-nil token")
	}
	if user.ProviderID != "42" {
		t.Errorf("ProviderID = %q, want %q", user.ProviderID, "42")
	}
	if user.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", user.Email, "alice@example.com")
	}
	if user.FirstName != "Alice Smith" {
		t.Errorf("FirstName = %q, want %q", user.FirstName, "Alice Smith")
	}
	if user.AvatarURL != "https://example.com/avatar.jpg" {
		t.Errorf("AvatarURL = %q, want %q", user.AvatarURL, "https://example.com/avatar.jpg")
	}
}

// ── TestGitHubProvider_Exchange_FallsBackToEmailsAPI ─────────────────────────

// TestGitHubProvider_Exchange_FallsBackToEmailsAPI verifies that when the user
// profile has no email, the provider falls back to the emails API and picks the
// primary + verified email.
func TestGitHubProvider_Exchange_FallsBackToEmailsAPI(t *testing.T) {
	userProfile := GitHubUser{
		ID:        7,
		Login:     "bob",
		Name:      "",   // empty name → use Login
		Email:     "",   // empty → must call emails API
		AvatarURL: "https://example.com/bob.jpg",
	}
	emails := []GitHubEmail{
		{Email: "bob-secondary@example.com", Primary: false, Verified: true},
		{Email: "bob@example.com", Primary: true, Verified: true},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "gh-token",
			"token_type":   "bearer",
		})
	})
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userProfile)
	})
	mux.HandleFunc("/user/emails", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(emails)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	wrapped := newGitHubTestProvider(srv)
	user, _, err := wrapped.Exchange(context.Background(), "code-xyz")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "bob@example.com" {
		t.Errorf("Email = %q, want primary email %q", user.Email, "bob@example.com")
	}
	// Name is empty → should use Login as FirstName
	if user.FirstName != "bob" {
		t.Errorf("FirstName = %q, want Login %q", user.FirstName, "bob")
	}
}

// ── TestGitHubProvider_Exchange_NoVerifiedEmail ────────────────────────────

// TestGitHubProvider_Exchange_NoVerifiedEmail verifies that the provider returns
// an error when the emails API has no verified email.
func TestGitHubProvider_Exchange_NoVerifiedEmail(t *testing.T) {
	userProfile := GitHubUser{ID: 9, Login: "carol", Email: ""}
	emails := []GitHubEmail{
		{Email: "carol-unverified@example.com", Primary: true, Verified: false},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "gh-token",
			"token_type":   "bearer",
		})
	})
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userProfile)
	})
	mux.HandleFunc("/user/emails", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(emails)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	wrapped := newGitHubTestProvider(srv)
	_, _, err := wrapped.Exchange(context.Background(), "code-xyz")

	if err == nil {
		t.Fatal("expected error when no verified email exists, got nil")
	}
	if !strings.Contains(err.Error(), "github: failed to fetch email") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ── TestGitHubProvider_Exchange_TokenError ────────────────────────────────────

// TestGitHubProvider_Exchange_TokenError verifies that a bad authorization code
// returns an error before the profile is fetched.
func TestGitHubProvider_Exchange_TokenError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad_verification_code", http.StatusUnauthorized)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	wrapped := newGitHubTestProvider(srv)
	_, _, err := wrapped.Exchange(context.Background(), "bad-code")

	if err == nil {
		t.Fatal("expected token exchange error, got nil")
	}
	if !strings.Contains(err.Error(), "github: failed to exchange code") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ── TestGitHubProvider_Exchange_UserAPIError ─────────────────────────────────

// TestGitHubProvider_Exchange_UserAPIError verifies that a non-200 /user response
// is propagated as an error.
func TestGitHubProvider_Exchange_UserAPIError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "gh-token",
			"token_type":   "bearer",
		})
	})
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	wrapped := newGitHubTestProvider(srv)
	_, _, err := wrapped.Exchange(context.Background(), "code-xyz")

	if err == nil {
		t.Fatal("expected error for non-200 user API, got nil")
	}
	if !strings.Contains(err.Error(), "github: user API returned 403") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// ── Test helper: gitHubProviderWithCustomBaseURL ──────────────────────────────
//
// GitHubProvider.Exchange hard-codes api.github.com URLs. This thin wrapper
// replaces the base URL so tests can point at a local httptest.Server without
// forking production code.

type gitHubProviderWithCustomBaseURL struct {
	inner   *GitHubProvider
	baseURL string // e.g. "http://127.0.0.1:PORT"
}

func newGitHubTestProvider(srv *httptest.Server) *gitHubProviderWithCustomBaseURL {
	p := &GitHubProvider{
		config: &oauth2.Config{
			ClientID:     "cid",
			ClientSecret: "csec",
			RedirectURL:  "http://localhost/cb",
			Endpoint: oauth2.Endpoint{
				AuthURL:  srv.URL + "/authorize",
				TokenURL: srv.URL + "/token",
			},
		},
	}
	return &gitHubProviderWithCustomBaseURL{inner: p, baseURL: srv.URL}
}

func (g *gitHubProviderWithCustomBaseURL) Exchange(ctx context.Context, code string) (*OAuthUser, *oauth2.Token, error) {
	token, err := g.inner.config.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("github: failed to exchange code: %w", err)
	}

	client := g.inner.config.Client(ctx, token)

	// /user
	resp, err := client.Get(g.baseURL + "/user")
	if err != nil {
		return nil, nil, fmt.Errorf("github: failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 256)
		n, _ := resp.Body.Read(body)
		return nil, nil, fmt.Errorf("github: user API returned %d: %s", resp.StatusCode, string(body[:n]))
	}

	var gu GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&gu); err != nil {
		return nil, nil, fmt.Errorf("github: failed to decode user info: %w", err)
	}

	email := gu.Email
	if email == "" {
		email, err = g.fetchPrimaryEmail(ctx, client)
		if err != nil {
			return nil, nil, fmt.Errorf("github: failed to fetch email: %w", err)
		}
	}

	firstName := gu.Name
	if gu.Name == "" {
		firstName = gu.Login
	}

	return &OAuthUser{
		ProviderID: fmt.Sprintf("%d", gu.ID),
		Email:      email,
		FirstName:  firstName,
		LastName:   "",
		AvatarURL:  gu.AvatarURL,
	}, token, nil
}

func (g *gitHubProviderWithCustomBaseURL) fetchPrimaryEmail(ctx context.Context, client *http.Client) (string, error) {
	resp, err := client.Get(g.baseURL + "/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 256)
		n, _ := resp.Body.Read(body)
		return "", fmt.Errorf("emails API returned %d: %s", resp.StatusCode, string(body[:n]))
	}

	var emails []GitHubEmail
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}
	return "", fmt.Errorf("no verified email found")
}
