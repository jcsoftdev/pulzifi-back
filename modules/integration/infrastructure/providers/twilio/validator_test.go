package twilioprovider_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	twilioprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/twilio"
)

func TestValidate_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2010-04-01/Accounts/AC123.json" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		expected := "Basic " + base64.StdEncoding.EncodeToString([]byte("AC123:tok"))
		if r.Header.Get("Authorization") != expected {
			t.Errorf("unexpected Authorization header: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"sid":"AC123","status":"active"}`)
	}))
	t.Cleanup(func() { srv.Close() })

	orig := twilioprovider.APIBase
	twilioprovider.APIBase = srv.URL
	t.Cleanup(func() { twilioprovider.APIBase = orig })

	v := twilioprovider.NewValidator()
	err := v.Validate(context.Background(), "twilio", map[string]string{
		"account_sid": "AC123",
		"auth_token":  "tok",
		"from_number": "+15005550006",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidate_BadCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	t.Cleanup(func() { srv.Close() })

	orig := twilioprovider.APIBase
	twilioprovider.APIBase = srv.URL
	t.Cleanup(func() { twilioprovider.APIBase = orig })

	v := twilioprovider.NewValidator()
	err := v.Validate(context.Background(), "twilio", map[string]string{
		"account_sid": "AC123",
		"auth_token":  "bad",
		"from_number": "+15005550006",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid credentials") {
		t.Errorf("expected 'invalid credentials' in error, got: %v", err)
	}
}

func TestValidate_Forbidden(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(func() { srv.Close() })

	orig := twilioprovider.APIBase
	twilioprovider.APIBase = srv.URL
	t.Cleanup(func() { twilioprovider.APIBase = orig })

	v := twilioprovider.NewValidator()
	err := v.Validate(context.Background(), "twilio", map[string]string{
		"account_sid": "AC123",
		"auth_token":  "tok",
		"from_number": "+15005550006",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid credentials") {
		t.Errorf("expected 'invalid credentials' in error, got: %v", err)
	}
}

func TestValidate_AccountSuspended(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"sid":"AC123","status":"suspended"}`)
	}))
	t.Cleanup(func() { srv.Close() })

	orig := twilioprovider.APIBase
	twilioprovider.APIBase = srv.URL
	t.Cleanup(func() { twilioprovider.APIBase = orig })

	v := twilioprovider.NewValidator()
	err := v.Validate(context.Background(), "twilio", map[string]string{
		"account_sid": "AC123",
		"auth_token":  "tok",
		"from_number": "+15005550006",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "status=suspended") {
		t.Errorf("expected 'status=suspended' in error, got: %v", err)
	}
}

func TestValidate_TwilioError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, `{"message":"Service unavailable"}`)
	}))
	t.Cleanup(func() { srv.Close() })

	orig := twilioprovider.APIBase
	twilioprovider.APIBase = srv.URL
	t.Cleanup(func() { twilioprovider.APIBase = orig })

	v := twilioprovider.NewValidator()
	err := v.Validate(context.Background(), "twilio", map[string]string{
		"account_sid": "AC123",
		"auth_token":  "tok",
		"from_number": "+15005550006",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Errorf("expected '503' in error, got: %v", err)
	}
}

func TestValidate_MissingSID(t *testing.T) {
	var hit bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(func() { srv.Close() })

	orig := twilioprovider.APIBase
	twilioprovider.APIBase = srv.URL
	t.Cleanup(func() { twilioprovider.APIBase = orig })

	v := twilioprovider.NewValidator()
	err := v.Validate(context.Background(), "twilio", map[string]string{
		"auth_token":  "x",
		"from_number": "+15005550006",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "missing account_sid") {
		t.Errorf("expected 'missing account_sid' in error, got: %v", err)
	}
	if hit {
		t.Error("expected no HTTP call to be made")
	}
}

func TestValidate_MissingAuthToken(t *testing.T) {
	var hit bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(func() { srv.Close() })

	orig := twilioprovider.APIBase
	twilioprovider.APIBase = srv.URL
	t.Cleanup(func() { twilioprovider.APIBase = orig })

	v := twilioprovider.NewValidator()
	err := v.Validate(context.Background(), "twilio", map[string]string{
		"account_sid": "AC123",
		"from_number": "+15005550006",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "missing auth_token") {
		t.Errorf("expected 'missing auth_token' in error, got: %v", err)
	}
	if hit {
		t.Error("expected no HTTP call to be made")
	}
}

func TestValidate_MissingFromNumber(t *testing.T) {
	var hit bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(func() { srv.Close() })

	orig := twilioprovider.APIBase
	twilioprovider.APIBase = srv.URL
	t.Cleanup(func() { twilioprovider.APIBase = orig })

	v := twilioprovider.NewValidator()
	err := v.Validate(context.Background(), "twilio", map[string]string{
		"account_sid": "AC123",
		"auth_token":  "tok",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "missing from_number") {
		t.Errorf("expected 'missing from_number' in error, got: %v", err)
	}
	if hit {
		t.Error("expected no HTTP call to be made")
	}
}

func TestValidate_UnsupportedProvider(t *testing.T) {
	v := twilioprovider.NewValidator()
	err := v.Validate(context.Background(), "slack", map[string]string{
		"account_sid": "AC123",
		"auth_token":  "tok",
		"from_number": "+15005550006",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported provider") {
		t.Errorf("expected 'unsupported provider' in error, got: %v", err)
	}
}
