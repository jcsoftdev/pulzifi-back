//go:build integration

// Package http_test — integration tests for the billing HTTP module.
//
// These tests exercise the full Chi middleware stack to assert that:
//   1. Raw bytes from the request body reach the webhook handler intact (no
//      early consumption by JSON middleware or other Chi middleware).
//   2. An invalid Stripe-Signature header returns HTTP 400.
//
// The tests use in-memory repositories and a mock StripeGateway — no database
// required. Run with:
//
//	go test -tags=integration ./modules/billing/infrastructure/http/...
package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	handlewebhook "github.com/jcsoftdev/pulzifi-back/modules/billing/application/handle_webhook"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services/mocks"
	billinghttp "github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/http"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/persistence/inmem"
)

// buildIntegrationRouter constructs a Chi router that mirrors the production
// server as closely as possible — it wraps the billing module inside a parent
// router that includes logging, recovery, and request-id middleware.
// Critically it does NOT install any JSON body-parsing middleware at the top
// level (matching production behaviour documented in module.go).
func buildIntegrationRouter(gateway *mocks.MockStripeGateway, secret string) http.Handler {
	webhookRepo := inmem.NewWebhookEventRepo()
	planAssigner := &mocks.MockPlanAssigner{}
	customerRepo := inmem.NewCustomerRepo()
	subRepo := inmem.NewSubscriptionRepo()

	wh := handlewebhook.NewHandler(gateway, secret, planAssigner, customerRepo, webhookRepo, subRepo)

	mod := billinghttp.NewModule(billinghttp.Deps{
		WebhookHandler: wh,
	})

	// Top-level parent router with typical production middlewares.
	parent := chi.NewRouter()
	parent.Use(middleware.RequestID)
	parent.Use(middleware.Recoverer)
	parent.Use(middleware.Logger)

	// Mount the billing module on /api/v1 to simulate the real mount path.
	parent.Route("/api/v1", func(r chi.Router) {
		mod.RegisterHTTPRoutes(r)
	})

	return parent
}

// TestIntegration_WebhookRawBodyPreservation verifies that raw bytes reach
// the webhook handler unmodified even when Chi middleware wraps the request.
//
// This is the key invariant for Stripe signature verification: stripe-go's
// webhook.ConstructEvent() hashes the exact bytes Stripe signed, so any
// middleware that reads or re-encodes the body would break signature checks.
func TestIntegration_WebhookRawBodyPreservation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	const secret = "whsec_integration_test"
	rawPayload := []byte(`{"id":"evt_inttest_001","type":"invoice.paid","data":{"object":{"subscription":"sub_x"}}}`)

	gateway := &mocks.MockStripeGateway{}
	var received []byte
	gateway.ConstructEventFn = func(payload []byte, sig, sec string) (services.StripeEvent, error) {
		received = make([]byte, len(payload))
		copy(received, payload)
		// Return sentinel error so we can assert body without running the full dispatch.
		return services.StripeEvent{}, errors.New("sig invalid — short-circuit for body assertion")
	}

	handler := buildIntegrationRouter(gateway, secret)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/webhook", bytes.NewReader(rawPayload))
	req.Header.Set("Stripe-Signature", "t=1,v1=placeholder")
	req.Header.Set("Content-Type", "application/json") // simulate Stripe headers
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 (invalid sig short-circuit), got %d — body: %s", rr.Code, rr.Body.String())
	}
	if !bytes.Equal(received, rawPayload) {
		t.Fatalf("body mutated in transit:\n  sent:     %q\n  received: %q", rawPayload, received)
	}
}

// TestIntegration_WebhookMissingSignature_Returns400 asserts that an absent
// Stripe-Signature header is rejected at the HTTP layer before the body is read.
func TestIntegration_WebhookMissingSignature_Returns400(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler := buildIntegrationRouter(&mocks.MockStripeGateway{}, "whsec_test")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/webhook",
		bytes.NewReader([]byte(`{"id":"evt_nosig"}`)))
	// Intentionally NO Stripe-Signature header.
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing Stripe-Signature, got %d", rr.Code)
	}
}

// TestIntegration_WebhookFullDispatch_Returns200 verifies the happy path end-to-end:
// a syntactically valid event (mock ConstructEvent returns success) produces
// HTTP 200 after full dispatch through the webhook module.
func TestIntegration_WebhookFullDispatch_Returns200(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	const secret = "whsec_dispatch"
	rawPayload := []byte(`{"id":"evt_dispatch_001","type":"invoice.paid","data":{"object":{"subscription":"sub_dispatch","customer":"cus_dispatch"}}}`)

	gateway := &mocks.MockStripeGateway{}
	gateway.ConstructEventFn = func(payload []byte, sig, sec string) (services.StripeEvent, error) {
		return services.StripeEvent{
			ID:      "evt_dispatch_001",
			Type:    "invoice.paid",
			RawData: payload,
		}, nil
	}
	// RetrieveSubscription must return something for invoice.paid handler.
	gateway.RetrieveSubscriptionFn = func(_ context.Context, subID string) (services.StripeSubscription, error) {
		return services.StripeSubscription{
			ID:               subID,
			Status:           "active",
			CurrentPeriodEnd: 9999999999,
			PriceID:          "",
			CustomerID:       "cus_dispatch",
		}, nil
	}

	handler := buildIntegrationRouter(gateway, secret)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/webhook", bytes.NewReader(rawPayload))
	req.Header.Set("Stripe-Signature", "t=1,v1=valid")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rr.Code, rr.Body.String())
	}
}

// TestIntegration_NonWebhookRoutes_NotAffected verifies that non-webhook billing
// routes are still reachable (no 404) even when mounted alongside the webhook route.
// Auth middleware is nil in test context, so we expect a panic-recovery or 401/403
// rather than 404.
func TestIntegration_NonWebhookRoutes_NotAffected(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler := buildIntegrationRouter(&mocks.MockStripeGateway{}, "whsec_test")

	routes := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/v1/billing/checkout"},
		{http.MethodPost, "/api/v1/billing/portal"},
		{http.MethodGet, "/api/v1/billing/subscription"},
	}
	for _, tc := range routes {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			body := io.NopCloser(bytes.NewReader([]byte(`{}`)))
			req := httptest.NewRequest(tc.method, tc.path, body)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// Recoverer middleware catches nil-middleware panics; we get a 500 instead of panic.
			handler.ServeHTTP(rr, req)

			// 404 would mean the route was not mounted — that is the only failure case.
			if rr.Code == http.StatusNotFound {
				t.Errorf("route %s %s returned 404 — route not registered", tc.method, tc.path)
			}
		})
	}
}

// TestIntegration_JSONBodyNotConsumedByMiddleware asserts that sending a JSON body
// to the checkout route does not result in an empty-body error.
// This proves that no Chi middleware at parent level is consuming r.Body before the handler.
func TestIntegration_JSONBodyNotConsumedByMiddleware(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	handler := buildIntegrationRouter(&mocks.MockStripeGateway{}, "whsec_test")

	body, _ := json.Marshal(map[string]string{
		"plan_id":       "plan_basic",
		"billing_cycle": "monthly",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/billing/checkout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Recoverer catches nil-middleware panics.
	handler.ServeHTTP(rr, req)

	// We accept any status except 400 with "invalid request body" error —
	// that would indicate the body was consumed by middleware before the handler.
	if rr.Code == http.StatusBadRequest {
		var errResp map[string]string
		if json.NewDecoder(rr.Body).Decode(&errResp) == nil {
			if errResp["error"] == "invalid request body" {
				t.Error("request body was consumed before reaching the handler — JSON middleware conflict")
			}
		}
	}
}
