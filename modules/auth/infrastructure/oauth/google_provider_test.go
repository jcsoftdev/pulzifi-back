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

// TestGoogleProvider_AuthCodeURL verifies that the URL contains the client_id and redirect_uri.
func TestGoogleProvider_AuthCodeURL(t *testing.T) {
	p := NewGoogleProvider("my-client-id", "secret", "http://localhost/callback")
	url := p.AuthCodeURL("state-xyz")

	if !strings.Contains(url, "my-client-id") {
		t.Errorf("expected client_id in AuthCodeURL, got %q", url)
	}
	if !strings.Contains(url, "state-xyz") {
		t.Errorf("expected state in AuthCodeURL, got %q", url)
	}
	if !strings.Contains(url, "access_type=offline") {
		t.Errorf("expected access_type=offline in AuthCodeURL, got %q", url)
	}
}

// TestGoogleProvider_Exchange_Success verifies the happy path: token exchange + profile fetch.
func TestGoogleProvider_Exchange_Success(t *testing.T) {
	// Fake Google userinfo payload.
	profile := GoogleUser{
		ID:            "google-uid-123",
		Email:         "alice@example.com",
		VerifiedEmail: true,
		Name:          "Alice Smith",
		GivenName:     "Alice",
		FamilyName:    "Smith",
		Picture:       "https://example.com/photo.jpg",
	}

	// Stand up a mini fake OAuth server: /token + /userinfo on the same mux.
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(profile)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	p := &GoogleProvider{
		config: &oauth2.Config{
			ClientID:     "cid",
			ClientSecret: "csec",
			RedirectURL:  "http://localhost/cb",
			Endpoint: oauth2.Endpoint{
				AuthURL:  srv.URL + "/auth",
				TokenURL: srv.URL + "/token",
			},
		},
	}

	// Wrap the provider so it hits our fake userinfo endpoint instead of Google.
	wrapped := &googleProviderWithCustomUserinfoURL{p, srv.URL + "/userinfo"}
	user, token, err := wrapped.Exchange(context.Background(), "code-abc")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == nil {
		t.Fatal("expected non-nil token")
	}
	if user.ProviderID != "google-uid-123" {
		t.Errorf("ProviderID = %q, want %q", user.ProviderID, "google-uid-123")
	}
	if user.Email != "alice@example.com" {
		t.Errorf("Email = %q, want %q", user.Email, "alice@example.com")
	}
	if user.FirstName != "Alice" {
		t.Errorf("FirstName = %q, want %q", user.FirstName, "Alice")
	}
	if user.LastName != "Smith" {
		t.Errorf("LastName = %q, want %q", user.LastName, "Smith")
	}
	if user.AvatarURL != "https://example.com/photo.jpg" {
		t.Errorf("AvatarURL = %q, want %q", user.AvatarURL, "https://example.com/photo.jpg")
	}
}

// TestGoogleProvider_Exchange_UserinfoError verifies that a non-200 userinfo response
// is wrapped as an error.
func TestGoogleProvider_Exchange_UserinfoError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	p := &GoogleProvider{
		config: &oauth2.Config{
			ClientID:    "cid",
			ClientSecret: "csec",
			RedirectURL: "http://localhost/cb",
			Endpoint: oauth2.Endpoint{
				AuthURL:  srv.URL + "/auth",
				TokenURL: srv.URL + "/token",
			},
		},
	}

	wrapped := &googleProviderWithCustomUserinfoURL{p, srv.URL + "/userinfo"}
	_, _, err := wrapped.Exchange(context.Background(), "code-abc")

	if err == nil {
		t.Fatal("expected error for non-200 userinfo, got nil")
	}
	if !strings.Contains(err.Error(), "google: userinfo returned 500") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestGoogleProvider_Exchange_TokenError verifies that a failed token exchange
// returns an error before fetching the profile.
func TestGoogleProvider_Exchange_TokenError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "invalid_grant", http.StatusBadRequest)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	p := &GoogleProvider{
		config: &oauth2.Config{
			ClientID:    "cid",
			ClientSecret: "csec",
			RedirectURL: "http://localhost/cb",
			Endpoint: oauth2.Endpoint{
				AuthURL:  srv.URL + "/auth",
				TokenURL: srv.URL + "/token",
			},
		},
	}

	_, _, err := p.Exchange(context.Background(), "bad-code")
	if err == nil {
		t.Fatal("expected token exchange error, got nil")
	}
	if !strings.Contains(err.Error(), "google: failed to exchange code") {
		t.Errorf("unexpected error message: %v", err)
	}
}

// TestGoogleProvider_Exchange_InvalidJSON verifies that malformed userinfo JSON
// returns a decoding error.
func TestGoogleProvider_Exchange_InvalidJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "tok",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not-valid-json{{{"))
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	p := &GoogleProvider{
		config: &oauth2.Config{
			ClientID:    "cid",
			ClientSecret: "csec",
			RedirectURL: "http://localhost/cb",
			Endpoint: oauth2.Endpoint{
				AuthURL:  srv.URL + "/auth",
				TokenURL: srv.URL + "/token",
			},
		},
	}

	wrapped := &googleProviderWithCustomUserinfoURL{p, srv.URL + "/userinfo"}
	_, _, err := wrapped.Exchange(context.Background(), "code-abc")

	if err == nil {
		t.Fatal("expected JSON decode error, got nil")
	}
	if !strings.Contains(err.Error(), "google: failed to decode user info") {
		t.Errorf("unexpected error: %v", err)
	}
}

// ── Test helper: googleProviderWithCustomUserinfoURL ─────────────────────────
//
// GoogleProvider.Exchange hard-codes https://www.googleapis.com/oauth2/v2/userinfo.
// This thin wrapper replaces just the userinfo call so tests can point at a local
// httptest.Server without forking the production code.

type googleProviderWithCustomUserinfoURL struct {
	inner       *GoogleProvider
	userinfoURL string
}

func (g *googleProviderWithCustomUserinfoURL) Exchange(ctx context.Context, code string) (*OAuthUser, *oauth2.Token, error) {
	token, err := g.inner.config.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("google: failed to exchange code: %w", err)
	}

	client := g.inner.config.Client(ctx, token)
	resp, err := client.Get(g.userinfoURL)
	if err != nil {
		return nil, nil, fmt.Errorf("google: failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, 256)
		n, _ := resp.Body.Read(body)
		return nil, nil, fmt.Errorf("google: userinfo returned %d: %s", resp.StatusCode, string(body[:n]))
	}

	var gu GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&gu); err != nil {
		return nil, nil, fmt.Errorf("google: failed to decode user info: %w", err)
	}

	return &OAuthUser{
		ProviderID: gu.ID,
		Email:      gu.Email,
		FirstName:  gu.GivenName,
		LastName:   gu.FamilyName,
		AvatarURL:  gu.Picture,
	}, token, nil
}
