package oauth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

var testKey = bytes.Repeat([]byte{0xAB}, 32)

func newSigner(ttl time.Duration) *StateSigner {
	return NewStateSigner(testKey, ttl)
}

func fullClaims() StateClaims {
	return StateClaims{
		Provider:    "github",
		Tenant:      "acme",
		OrgID:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		UserID:      uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		ReturnPath:  "/dashboard",
		RedirectURI: "https://acme.app.com/callback",
		Nonce:       "explicit-nonce",
	}
}

func TestSign_Verify_RoundTrip(t *testing.T) {
	s := newSigner(time.Hour)
	original := fullClaims()

	token, err := s.Sign(original)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	got, err := s.Verify(token)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}

	if got.Provider != original.Provider {
		t.Errorf("Provider mismatch: got %q, want %q", got.Provider, original.Provider)
	}
	if got.Tenant != original.Tenant {
		t.Errorf("Tenant mismatch: got %q, want %q", got.Tenant, original.Tenant)
	}
	if got.OrgID != original.OrgID {
		t.Errorf("OrgID mismatch: got %v, want %v", got.OrgID, original.OrgID)
	}
	if got.UserID != original.UserID {
		t.Errorf("UserID mismatch: got %v, want %v", got.UserID, original.UserID)
	}
	if got.ReturnPath != original.ReturnPath {
		t.Errorf("ReturnPath mismatch: got %q, want %q", got.ReturnPath, original.ReturnPath)
	}
	if got.RedirectURI != original.RedirectURI {
		t.Errorf("RedirectURI mismatch: got %q, want %q", got.RedirectURI, original.RedirectURI)
	}
	if got.Nonce != original.Nonce {
		t.Errorf("Nonce mismatch: got %q, want %q", got.Nonce, original.Nonce)
	}

	// Test auto-generated nonce: sign with empty Nonce
	empty := StateClaims{}
	token2, err := s.Sign(empty)
	if err != nil {
		t.Fatalf("Sign (empty nonce) failed: %v", err)
	}
	got2, err := s.Verify(token2)
	if err != nil {
		t.Fatalf("Verify (empty nonce) failed: %v", err)
	}
	// Nonce should be a valid UUID
	if _, err := uuid.Parse(got2.Nonce); err != nil {
		t.Errorf("Auto-generated nonce is not a valid UUID: %q", got2.Nonce)
	}
}

func TestVerify_RejectsTamperedHeader(t *testing.T) {
	s := newSigner(time.Hour)
	token, err := s.Sign(fullClaims())
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	// Flip the first byte of the header
	runes := []byte(token)
	runes[0] ^= 0xFF
	tampered := string(runes)

	_, err = s.Verify(tampered)
	if err == nil {
		t.Fatal("expected error for tampered header, got nil")
	}
	if err.Error() != "bad signature" {
		t.Errorf("expected 'bad signature', got %q", err.Error())
	}
}

func TestVerify_RejectsExpiredToken(t *testing.T) {
	// Use a negative TTL so ExpiresAt is already in the past at sign time.
	s := newSigner(-time.Second)
	token, err := s.Sign(fullClaims())
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	_, err = s.Verify(token)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
	if err.Error() != "state expired" {
		t.Errorf("expected 'state expired', got %q", err.Error())
	}
}

func TestVerify_RejectsBadSignature(t *testing.T) {
	s := newSigner(time.Hour)
	token, err := s.Sign(fullClaims())
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	dot := strings.Index(token, ".")
	head := token[:dot]
	fakeSig := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0x00}, 32))
	tampered := head + "." + fakeSig

	_, err = s.Verify(tampered)
	if err == nil {
		t.Fatal("expected error for bad signature, got nil")
	}
	if err.Error() != "bad signature" {
		t.Errorf("expected 'bad signature', got %q", err.Error())
	}
}

func TestSign_GeneratesUniqueNonce(t *testing.T) {
	s := newSigner(time.Hour)

	token1, err := s.Sign(StateClaims{})
	if err != nil {
		t.Fatalf("Sign 1 failed: %v", err)
	}
	token2, err := s.Sign(StateClaims{})
	if err != nil {
		t.Fatalf("Sign 2 failed: %v", err)
	}

	decodeNonce := func(token string) string {
		dot := strings.Index(token, ".")
		body, err := base64.RawURLEncoding.DecodeString(token[:dot])
		if err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		var c StateClaims
		if err := json.Unmarshal(body, &c); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		return c.Nonce
	}

	n1 := decodeNonce(token1)
	n2 := decodeNonce(token2)
	if n1 == n2 {
		t.Errorf("expected unique nonces, both are %q", n1)
	}
}

func TestVerify_RejectsMalformedToken(t *testing.T) {
	s := newSigner(time.Hour)

	cases := []struct {
		name  string
		token string
	}{
		{"no-dot", "nodotinhere"},
		{"empty", ""},
		{"only-sig", ".onlysig"},
		{"only-head-trailing-dot", "onlyhead."},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.Verify(tc.token)
			if err == nil {
				t.Errorf("expected non-nil error for token %q, got nil", tc.token)
			}
		})
	}
}
