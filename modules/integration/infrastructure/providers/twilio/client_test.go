package twilioprovider_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/integration/domain/entities"
	twilioprovider "github.com/jcsoftdev/pulzifi-back/modules/integration/infrastructure/providers/twilio"
)

// stubPlans is a test double for OrgPlanLookup.
type stubPlans struct {
	code string
	err  error
}

func (s *stubPlans) PlanCode(_ context.Context, _ uuid.UUID) (string, error) {
	return s.code, s.err
}

// newPaidClient builds a Client configured for paid-tier testing.
func newPaidClient(plans *stubPlans) *twilioprovider.Client {
	return twilioprovider.New(twilioprovider.Config{
		PaidPlans:          []string{"pro", "business"},
		PlatformAccountSID: "ACplatform",
		PlatformAuthToken:  "platformToken",
		PlatformFromNumber: "+10000000000",
	}, plans)
}

// orgDest returns a Destination scoped to an org.
func orgDest(phoneNumbers ...string) *entities.Destination {
	nums := make([]any, len(phoneNumbers))
	for i, n := range phoneNumbers {
		nums[i] = n
	}
	return &entities.Destination{
		ScopeType: entities.ScopeOrg,
		ScopeID:   uuid.New(),
		Target:    map[string]any{"phone_numbers": nums},
	}
}

func payload() *entities.NotificationPayload {
	return &entities.NotificationPayload{
		Title: "Alert",
		Body:  "Something happened",
	}
}

// newMockServer creates a test HTTP server with a customizable handler function.
func newMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(func() { srv.Close() })
	return srv
}

// swapAPIBase sets APIBase to the given URL and restores it on cleanup.
func swapAPIBase(t *testing.T, base string) {
	t.Helper()
	orig := twilioprovider.APIBase
	twilioprovider.APIBase = base
	t.Cleanup(func() { twilioprovider.APIBase = orig })
}

// ─── OAuth no-op tests ────────────────────────────────────────────────────────

func TestTwilio_OAuthIsNoOp(t *testing.T) {
	c := twilioprovider.New(twilioprovider.Config{}, &stubPlans{})
	ctx := context.Background()

	// OAuthAuthorizeURL returns empty string and non-nil error.
	u, err := c.OAuthAuthorizeURL("state", "redirect")
	if u != "" {
		t.Errorf("expected empty URL, got %q", u)
	}
	if err == nil {
		t.Error("expected non-nil error from OAuthAuthorizeURL")
	}

	// HandleCallback returns nil result and non-nil error.
	res, err := c.HandleCallback(ctx, "code", "redirect")
	if res != nil {
		t.Errorf("expected nil result, got %+v", res)
	}
	if err == nil {
		t.Error("expected non-nil error from HandleCallback")
	}

	// RefreshAccessToken returns nil result and nil error (no-op).
	oauthRes, err := c.RefreshAccessToken(ctx, "refresh")
	if oauthRes != nil {
		t.Errorf("expected nil result, got %+v", oauthRes)
	}
	if err != nil {
		t.Errorf("expected nil error from RefreshAccessToken, got: %v", err)
	}
}

// ─── Tier routing tests ───────────────────────────────────────────────────────

func TestTwilio_TierFor_Enterprise(t *testing.T) {
	// Enterprise tier when integ != nil, regardless of plan.
	var hit bool
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"sid":"SMenterprise","status":"queued"}`)
	})
	swapAPIBase(t, srv.URL)

	c := twilioprovider.New(twilioprovider.Config{
		PaidPlans: []string{"pro"},
		// No platform creds — enterprise must use integ creds only.
	}, &stubPlans{code: "free"}) // plan is "free" but integ != nil → enterprise

	integ := &entities.Integration{
		RefreshToken: "ACenterprise",
		AccessToken:  "enterpriseToken",
		ProviderMeta: map[string]any{"from_number": "+19999999999"},
	}

	result, err := c.Send(context.Background(), integ, orgDest("+12345678901"), payload())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !hit {
		t.Error("expected HTTP call to be made for enterprise tier")
	}
	if result == nil || result.Code != http.StatusCreated {
		t.Errorf("expected result.Code=201, got %+v", result)
	}
}

func TestTwilio_TierFor_Paid(t *testing.T) {
	var capturedAuth string
	var capturedFrom string
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		if err := r.ParseForm(); err == nil {
			capturedFrom = r.FormValue("From")
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"sid":"SMpaid","status":"queued"}`)
	})
	swapAPIBase(t, srv.URL)

	plans := &stubPlans{code: "pro"}
	c := newPaidClient(plans)

	result, err := c.Send(context.Background(), nil, orgDest("+12345678901"), payload())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil || result.Code != http.StatusCreated {
		t.Errorf("expected result.Code=201, got %+v", result)
	}

	expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("ACplatform:platformToken"))
	if capturedAuth != expectedAuth {
		t.Errorf("expected Authorization %q, got %q", expectedAuth, capturedAuth)
	}
	if capturedFrom != "+10000000000" {
		t.Errorf("expected From=+10000000000, got %q", capturedFrom)
	}
}

func TestTwilio_TierFor_Free(t *testing.T) {
	var hit bool
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(http.StatusCreated)
	})
	swapAPIBase(t, srv.URL)

	plans := &stubPlans{code: "starter"}
	c := twilioprovider.New(twilioprovider.Config{
		PaidPlans: []string{"pro"},
	}, plans)

	_, err := c.Send(context.Background(), nil, orgDest("+12345678901"), payload())
	if err == nil {
		t.Fatal("expected error for free tier, got nil")
	}
	if !strings.Contains(err.Error(), "SMS not available on free plan") {
		t.Errorf("expected 'SMS not available on free plan', got: %v", err)
	}
	if hit {
		t.Error("expected no HTTP call for free tier")
	}
}

func TestTwilio_TierFor_PlanLookupError(t *testing.T) {
	plans := &stubPlans{err: errors.New("db down")}
	c := twilioprovider.New(twilioprovider.Config{}, plans)

	_, err := c.Send(context.Background(), nil, orgDest("+12345678901"), payload())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "tier lookup") {
		t.Errorf("expected 'tier lookup' in error, got: %v", err)
	}
}

// ─── Send credential verification tests ──────────────────────────────────────

func TestTwilio_Send_PaidUsesPlatformCreds(t *testing.T) {
	var capturedAuth string
	var capturedFrom string
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		if err := r.ParseForm(); err == nil {
			capturedFrom = r.FormValue("From")
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"sid":"SMxxx","status":"queued"}`)
	})
	swapAPIBase(t, srv.URL)

	plans := &stubPlans{code: "pro"}
	c := newPaidClient(plans)

	result, err := c.Send(context.Background(), nil, orgDest("+12345678901"), payload())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil || result.Code != http.StatusCreated {
		t.Errorf("expected result.Code=201, got %+v", result)
	}

	expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("ACplatform:platformToken"))
	if capturedAuth != expectedAuth {
		t.Errorf("expected Authorization %q, got %q", expectedAuth, capturedAuth)
	}
	if capturedFrom != "+10000000000" {
		t.Errorf("expected From=+10000000000, got %q", capturedFrom)
	}
}

func TestTwilio_Send_EnterpriseUsesIntegCreds(t *testing.T) {
	var capturedAuth string
	var capturedFrom string
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		if err := r.ParseForm(); err == nil {
			capturedFrom = r.FormValue("From")
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"sid":"SMenterprise","status":"queued"}`)
	})
	swapAPIBase(t, srv.URL)

	c := twilioprovider.New(twilioprovider.Config{
		PaidPlans:          []string{"pro"},
		PlatformAccountSID: "ACplatform",
		PlatformAuthToken:  "platformToken",
		PlatformFromNumber: "+10000000000",
	}, &stubPlans{code: "pro"})

	integ := &entities.Integration{
		RefreshToken: "ACenterprise",
		AccessToken:  "enterpriseToken",
		ProviderMeta: map[string]any{"from_number": "+19999999999"},
	}

	result, err := c.Send(context.Background(), integ, orgDest("+12345678901"), payload())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil || result.Code != http.StatusCreated {
		t.Errorf("expected result.Code=201, got %+v", result)
	}

	expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("ACenterprise:enterpriseToken"))
	if capturedAuth != expectedAuth {
		t.Errorf("expected enterprise Authorization %q, got %q", expectedAuth, capturedAuth)
	}
	if capturedFrom != "+19999999999" {
		t.Errorf("expected From=+19999999999, got %q", capturedFrom)
	}
}

// ─── Multiple recipients tests ────────────────────────────────────────────────

func TestTwilio_Send_MultipleRecipients(t *testing.T) {
	var count int32
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"sid":"SMxxx","status":"queued"}`)
	})
	swapAPIBase(t, srv.URL)

	plans := &stubPlans{code: "pro"}
	c := newPaidClient(plans)

	result, err := c.Send(context.Background(), nil,
		orgDest("+11111111111", "+12222222222", "+13333333333"),
		payload(),
	)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if n := atomic.LoadInt32(&count); n != 3 {
		t.Errorf("expected 3 HTTP requests, got %d", n)
	}
}

func TestTwilio_Send_StopOnFirstFailure(t *testing.T) {
	var count int32
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&count, 1)
		if n == 1 {
			// First request succeeds.
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"sid":"SMxxx","status":"queued"}`)
			return
		}
		// Second request returns opted-out error (Twilio code 21610).
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, `{"code":21610,"message":"Message cannot be sent to the opted-out number"}`)
	})
	swapAPIBase(t, srv.URL)

	plans := &stubPlans{code: "pro"}
	c := newPaidClient(plans)

	_, err := c.Send(context.Background(), nil,
		orgDest("+11111111111", "+12222222222", "+13333333333"),
		payload(),
	)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if n := atomic.LoadInt32(&count); n != 2 {
		t.Errorf("expected exactly 2 HTTP requests, got %d", n)
	}
}

// ─── Error response handling tests ───────────────────────────────────────────

func TestTwilio_Send_AuthError_ReturnsCode401(t *testing.T) {
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"code":20003,"message":"Authenticate"}`)
	})
	swapAPIBase(t, srv.URL)

	plans := &stubPlans{code: "pro"}
	c := newPaidClient(plans)

	result, err := c.Send(context.Background(), nil, orgDest("+12345678901"), payload())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "auth") {
		t.Errorf("expected 'auth' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "20003") {
		t.Errorf("expected '20003' in error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result on auth error")
	}
	if result.Code != http.StatusUnauthorized {
		t.Errorf("expected result.Code=401, got %d", result.Code)
	}
}

func TestTwilio_Send_GenericError(t *testing.T) {
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, `{"code":30001,"message":"Queue overflow"}`)
	})
	swapAPIBase(t, srv.URL)

	plans := &stubPlans{code: "pro"}
	c := newPaidClient(plans)

	result, err := c.Send(context.Background(), nil, orgDest("+12345678901"), payload())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "30001") {
		t.Errorf("expected '30001' in error, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result on generic error")
	}
	if result.Code != http.StatusInternalServerError {
		t.Errorf("expected result.Code=500, got %d", result.Code)
	}
}

// ─── Validation / guard tests ─────────────────────────────────────────────────

func TestTwilio_Send_NoPhoneNumbers(t *testing.T) {
	var hit bool
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(http.StatusCreated)
	})
	swapAPIBase(t, srv.URL)

	plans := &stubPlans{code: "pro"}
	c := newPaidClient(plans)

	dest := &entities.Destination{
		ScopeType: entities.ScopeOrg,
		ScopeID:   uuid.New(),
		Target:    map[string]any{"phone_numbers": []any{}},
	}

	_, err := c.Send(context.Background(), nil, dest, payload())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no phone numbers") {
		t.Errorf("expected 'no phone numbers', got: %v", err)
	}
	if hit {
		t.Error("expected no HTTP call when no phone numbers")
	}
}

func TestTwilio_Send_OrgScopeRequired(t *testing.T) {
	var hit bool
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(http.StatusCreated)
	})
	swapAPIBase(t, srv.URL)

	plans := &stubPlans{code: "pro"}
	c := newPaidClient(plans)

	dest := &entities.Destination{
		ScopeType: entities.ScopeWorkspace,
		ScopeID:   uuid.New(),
		Target:    map[string]any{"phone_numbers": []any{"+12345678901"}},
	}

	_, err := c.Send(context.Background(), nil, dest, payload())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "tier lookup requires org-scoped destination") {
		t.Errorf("expected 'tier lookup requires org-scoped destination', got: %v", err)
	}
	if hit {
		t.Error("expected no HTTP call for non-org scope")
	}
}

func TestTwilio_Send_PaidWithoutPlatformCreds(t *testing.T) {
	var hit bool
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(http.StatusCreated)
	})
	swapAPIBase(t, srv.URL)

	// Paid tier but missing platform creds.
	plans := &stubPlans{code: "pro"}
	c := twilioprovider.New(twilioprovider.Config{
		PaidPlans:          []string{"pro"},
		PlatformAccountSID: "", // missing
		PlatformAuthToken:  "platformToken",
		PlatformFromNumber: "+10000000000",
	}, plans)

	_, err := c.Send(context.Background(), nil, orgDest("+12345678901"), payload())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "platform credentials not configured") {
		t.Errorf("expected 'platform credentials not configured', got: %v", err)
	}
	if hit {
		t.Error("expected no HTTP call when platform creds missing")
	}
}

func TestTwilio_Send_EnterpriseMissingFromNumber(t *testing.T) {
	var hit bool
	srv := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(http.StatusCreated)
	})
	swapAPIBase(t, srv.URL)

	c := twilioprovider.New(twilioprovider.Config{
		PaidPlans: []string{"pro"},
	}, &stubPlans{code: "pro"})

	// Enterprise integ with valid SID+token but missing from_number.
	integ := &entities.Integration{
		RefreshToken: "ACenterprise",
		AccessToken:  "enterpriseToken",
		ProviderMeta: map[string]any{
			// from_number intentionally absent
		},
	}

	_, err := c.Send(context.Background(), integ, orgDest("+12345678901"), payload())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "enterprise integration missing required fields") {
		t.Errorf("expected 'enterprise integration missing required fields', got: %v", err)
	}
	if hit {
		t.Error("expected no HTTP call when from_number missing")
	}
}
