package cookies

import (
	"net/http"
	"testing"
)

func TestGetTokenFromCookie_PrefersLastDuplicateCookie(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Cookie", "access_token=expired-old; access_token=fresh-new")

	token, err := GetTokenFromCookie(req, AccessTokenCookie)
	if err != nil {
		t.Fatalf("expected token without error, got: %v", err)
	}

	if token != "fresh-new" {
		t.Fatalf("expected latest cookie value 'fresh-new', got %q", token)
	}
}

func TestGetTokenFromCookie_MissingCookieReturnsError(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	_, getErr := GetTokenFromCookie(req, AccessTokenCookie)
	if getErr == nil {
		t.Fatal("expected error when cookie is missing")
	}
}
