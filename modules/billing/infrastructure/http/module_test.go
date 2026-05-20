package http_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	handlewebhook "github.com/jcsoftdev/pulzifi-back/modules/billing/application/handle_webhook"
	billinghttp "github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/http"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services/mocks"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/persistence/inmem"
)

// ── helpers ───────────────────────────────────────────────────────────────────

func buildWebhookModule(gateway *mocks.MockStripeGateway, secret string) (*billinghttp.Module, chi.Router) {
	webhookRepo := inmem.NewWebhookEventRepo()
	planAssigner := &mocks.MockPlanAssigner{}
	customerRepo := inmem.NewCustomerRepo()
	subRepo := inmem.NewSubscriptionRepo()

	wh := handlewebhook.NewHandler(gateway, secret, planAssigner, customerRepo, webhookRepo, subRepo)

	mod := billinghttp.NewModule(billinghttp.Deps{
		WebhookHandler: wh,
	})

	r := chi.NewRouter()
	mod.RegisterHTTPRoutes(r)
	return mod, r
}

// ── TestWebhookRawBody ────────────────────────────────────────────────────────

// TestWebhookRawBody verifies that the webhook handler receives the exact bytes
// that were sent — no JSON parsing middleware has consumed or altered the body.
//
// The mock StripeGateway.ConstructEventFn captures the payload it receives and
// returns ErrInvalidSignature to short-circuit (we only care that bytes arrived intact).
func TestWebhookRawBody(t *testing.T) {
	const secret = "whsec_testsecret"
	rawPayload := []byte(`{"id":"evt_test","type":"invoice.paid","data":{"object":{}}}`)

	gateway := &mocks.MockStripeGateway{}
	var capturedPayload []byte
	gateway.ConstructEventFn = func(payload []byte, signature, sec string) (services.StripeEvent, error) {
		capturedPayload = make([]byte, len(payload))
		copy(capturedPayload, payload)
		// Return error to short-circuit dispatch — we only care that payload arrived intact.
		return services.StripeEvent{}, errors.New("invalid sig")
	}

	_, r := buildWebhookModule(gateway, secret)

	req := httptest.NewRequest(http.MethodPost, "/billing/webhook", bytes.NewReader(rawPayload))
	req.Header.Set("Stripe-Signature", "t=123,v1=invalidsig") // non-empty but invalid
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	// Gateway returned an error → handler treats it as invalid signature → 400.
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (body: %s)", rr.Code, rr.Body.String())
	}

	// CRITICAL assertion: bytes must be identical.
	if !bytes.Equal(capturedPayload, rawPayload) {
		t.Fatalf("raw body was mutated: sent %q, gateway received %q", rawPayload, capturedPayload)
	}
}

// ── TestWebhookMissingSignature ───────────────────────────────────────────────

// TestWebhookMissingSignature verifies 400 is returned when Stripe-Signature header is absent.
// The HTTP layer must guard at this level before even reading the body.
func TestWebhookMissingSignature(t *testing.T) {
	_, r := buildWebhookModule(&mocks.MockStripeGateway{}, "whsec_test")

	req := httptest.NewRequest(http.MethodPost, "/billing/webhook", bytes.NewReader([]byte(`{}`)))
	// Intentionally NO Stripe-Signature header.
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing signature, got %d", rr.Code)
	}
}

// ── TestWebhookDuplicateEvent200 ─────────────────────────────────────────────

// TestWebhookDuplicateEvent200 verifies that a duplicate event returns HTTP 200.
// The use case idempotency (via in-mem webhook repo) returns nil on duplicate,
// so the handler must respond 200 OK — not an error.
func TestWebhookDuplicateEvent200(t *testing.T) {
	const secret = "whsec_testsecret"
	rawPayload := []byte(`{"id":"evt_dup","type":"invoice.paid","data":{"object":{}}}`)

	gateway := &mocks.MockStripeGateway{}
	// Return a minimal valid StripeEvent with known ID on every call.
	gateway.ConstructEventFn = func(payload []byte, sig, sec string) (services.StripeEvent, error) {
		return services.StripeEvent{
			ID:      "evt_dup",
			Type:    "invoice.paid",
			RawData: []byte(`{}`),
		}, nil
	}

	_, r := buildWebhookModule(gateway, secret)

	send := func() int {
		req := httptest.NewRequest(http.MethodPost, "/billing/webhook", bytes.NewReader(rawPayload))
		req.Header.Set("Stripe-Signature", "t=123,v1=valid")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		return rr.Code
	}

	// First call: event inserted → dispatcher runs → planAssigner mock returns nil → 200.
	code1 := send()
	// Second call: duplicate → Save returns (false, nil) → handler short-circuits → 200.
	code2 := send()

	if code1 != http.StatusOK {
		t.Errorf("first call: expected 200, got %d", code1)
	}
	if code2 != http.StatusOK {
		t.Errorf("duplicate event: expected 200, got %d", code2)
	}
}


// ── TestRoutesRegistered ──────────────────────────────────────────────────────

// TestRoutesRegistered verifies that all four billing routes are mounted on the router
// (non-404). Auth enforcement is tested in Phase 10 integration tests.
func TestRoutesRegistered(t *testing.T) {
	_, r := buildWebhookModule(&mocks.MockStripeGateway{}, "secret")

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/billing/webhook"},
		{http.MethodPost, "/billing/checkout"},
		{http.MethodPost, "/billing/portal"},
		{http.MethodGet, "/billing/subscription"},
	}

	for _, tc := range tests {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, http.NoBody)
			rr := httptest.NewRecorder()

			// Catch panics from nil auth middleware singleton in test context.
			panicked := false
			func() {
				defer func() {
					if rec := recover(); rec != nil {
						panicked = true
						t.Logf("nil-middleware panic (expected in unit test, not 404): %v", rec)
					}
				}()
				r.ServeHTTP(rr, req)
			}()

			// If no panic: any status != 404 means the route IS registered.
			if !panicked && rr.Code == http.StatusNotFound {
				t.Errorf("route %s %s returned 404 — route is not registered", tc.method, tc.path)
			}
		})
	}
}
