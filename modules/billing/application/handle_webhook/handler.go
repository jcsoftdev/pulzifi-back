package handlewebhook

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// ErrInvalidSignature is returned when the Stripe-Signature header is missing or invalid.
var ErrInvalidSignature = errors.New("billing: invalid or missing Stripe webhook signature")

// Handler orchestrates Stripe webhook processing.
type Handler struct {
	gateway        services.StripeGateway
	webhookSecret  string
	planAssigner   services.PlanAssigner
	trialConverter services.TrialConverter // optional; nil → trial conversion is skipped
	customerRepo   repositories.CustomerRepository
	webhookRepo    repositories.WebhookEventRepository
	subRepo        repositories.SubscriptionRepository
	planRepo       repositories.PlanRepository // optional; nil → pricing-sync events are skipped
}

// NewHandler returns a Handler with its dependencies injected.
func NewHandler(
	gateway services.StripeGateway,
	webhookSecret string,
	planAssigner services.PlanAssigner,
	customerRepo repositories.CustomerRepository,
	webhookRepo repositories.WebhookEventRepository,
	subRepo repositories.SubscriptionRepository,
) *Handler {
	return &Handler{
		gateway:       gateway,
		webhookSecret: webhookSecret,
		planAssigner:  planAssigner,
		customerRepo:  customerRepo,
		webhookRepo:   webhookRepo,
		subRepo:       subRepo,
	}
}

// WithTrialConverter wires the optional TrialConverter and returns the handler
// for chaining. When non-nil, the converter is invoked after a successful
// PlanAssigner.Assign on checkout.session.completed.
func (h *Handler) WithTrialConverter(tc services.TrialConverter) *Handler {
	h.trialConverter = tc
	return h
}

// WithPlanRepository wires the optional PlanRepository used to cache Stripe
// price amounts into public.plans on pricing webhook events. When nil, those
// events are acknowledged without writing.
func (h *Handler) WithPlanRepository(pr repositories.PlanRepository) *Handler {
	h.planRepo = pr
	return h
}

// Handle processes a raw Stripe webhook payload.
//
//  1. Reject empty signature header immediately.
//  2. ConstructEvent — validates HMAC signature.
//  3. Idempotency: insert event into stripe_webhook_events; skip if duplicate.
//  4. Dispatch to the appropriate sub-handler by event type.
//  5. Mark the event as processed.
func (h *Handler) Handle(ctx context.Context, rawBody []byte, signature string) error {
	// 1. Guard: missing signature
	if signature == "" {
		return ErrInvalidSignature
	}

	// 2. Verify signature and parse event
	event, err := h.gateway.ConstructEvent(rawBody, signature, h.webhookSecret)
	if err != nil {
		return ErrInvalidSignature
	}

	// 3. Idempotency check — persist the raw payload so deferred / failed events
	// can be replayed by ReconcileFromStripe without re-fetching from Stripe.
	we := &entities.WebhookEvent{
		EventID:    event.ID,
		EventType:  event.Type,
		ReceivedAt: time.Now(),
		Status:     entities.WebhookEventReceived,
		RawPayload: rawBody,
	}
	inserted, err := h.webhookRepo.Save(ctx, we)
	if err != nil {
		return err
	}
	if !inserted {
		// Duplicate — already processed; return 200 no-op
		return nil
	}

	// 4. Dispatch
	dispatchErr := h.dispatch(ctx, event)
	if dispatchErr != nil {
		// Orphan customer is NOT a failure — the org will appear later and
		// reconcile-on-link will replay the payload. Mark deferred and
		// acknowledge (do not propagate to HTTP layer).
		if errors.Is(dispatchErr, services.ErrOrphanCustomer) {
			logger.Warn("webhook deferred: customer not linked to any org yet",
				zap.String("event_id", event.ID),
				zap.String("event_type", event.Type),
			)
			_ = h.webhookRepo.MarkProcessed(ctx, event.ID, entities.WebhookEventDeferred)
			return nil
		}
		logger.Error("webhook dispatch error", zap.String("event_id", event.ID), zap.String("event_type", event.Type), zap.Error(dispatchErr))
		// Still mark as failed so ops can see it; don't return error to Stripe (would cause retry loop)
		_ = h.webhookRepo.MarkProcessed(ctx, event.ID, entities.WebhookEventFailed)
		return dispatchErr
	}

	// 5. Mark processed
	_ = h.webhookRepo.MarkProcessed(ctx, event.ID, entities.WebhookEventProcessed)
	return nil
}

// dispatch routes the event to the correct sub-handler.
func (h *Handler) dispatch(ctx context.Context, event services.StripeEvent) error {
	switch event.Type {
	case "checkout.session.completed":
		return h.handleCheckoutCompleted(ctx, event)
	case "invoice.paid":
		return h.handleInvoicePaid(ctx, event)
	case "customer.subscription.created", "customer.subscription.updated":
		return h.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		return h.handleSubscriptionDeleted(ctx, event)
	case "invoice.payment_failed":
		return h.handlePaymentFailed(ctx, event)
	case "price.created", "price.updated":
		return h.handlePricingChanged(ctx, event)
	default:
		// Unknown event — acknowledge silently (Stripe will not retry)
		logger.Info("billing: ignoring unknown webhook event type", zap.String("type", event.Type))
		return nil
	}
}

// ── Sub-handlers ──────────────────────────────────────────────────────────────

// checkoutSessionPayload is the subset we read from checkout.session.completed data.
type checkoutSessionPayload struct {
	Customer     string `json:"customer"`
	Subscription string `json:"subscription"`
}

// subscriptionPayload is the subset we read from subscription events.
//
// Stripe API ≥ 2024-12 moved `current_period_end` from the subscription
// top-level into each subscription item (items.data[].current_period_end).
// We read both and prefer the item-level value when present.
type subscriptionPayload struct {
	ID               string `json:"id"`
	Customer         string `json:"customer"`
	Status           string `json:"status"`
	CurrentPeriodEnd int64  `json:"current_period_end"`
	Items            struct {
		Data []struct {
			CurrentPeriodEnd int64 `json:"current_period_end"`
			Price            struct {
				ID string `json:"id"`
			} `json:"price"`
		} `json:"data"`
	} `json:"items"`
}

// resolveCurrentPeriodEnd returns the item-level current_period_end when the
// top-level value is missing (Stripe API ≥ 2024-12 behavior).
func (p subscriptionPayload) resolveCurrentPeriodEnd() int64 {
	if p.CurrentPeriodEnd > 0 {
		return p.CurrentPeriodEnd
	}
	if len(p.Items.Data) > 0 {
		return p.Items.Data[0].CurrentPeriodEnd
	}
	return 0
}

// invoicePayload is the subset we read from invoice events.
type invoicePayload struct {
	Customer     string `json:"customer"`
	Subscription string `json:"subscription"`
}

// pricePayload is the subset we read from price.created / price.updated events.
// A one-time (non-recurring) price has no recurring.interval and is skipped.
type pricePayload struct {
	ID         string `json:"id"`
	Product    string `json:"product"`
	UnitAmount int64  `json:"unit_amount"`
	Currency   string `json:"currency"`
	Recurring  *struct {
		Interval string `json:"interval"`
	} `json:"recurring"`
	Metadata struct {
		PlanCode string `json:"plan_code"`
	} `json:"metadata"`
}

func (h *Handler) handleCheckoutCompleted(ctx context.Context, event services.StripeEvent) error {
	var payload checkoutSessionPayload
	if err := json.Unmarshal(event.RawData, &payload); err != nil {
		return err
	}

	// Fetch full subscription from Stripe
	sub, err := h.gateway.RetrieveSubscription(ctx, payload.Subscription)
	if err != nil {
		return err
	}

	// Resolve org from customer
	orgID, err := h.resolveOrgByCustomer(ctx, payload.Customer)
	if err != nil {
		return err
	}

	// Persist customer mapping (may already exist)
	existing, _ := h.customerRepo.FindByOrgID(ctx, orgID)
	if existing == nil {
		_ = h.customerRepo.Save(ctx, &entities.Customer{
			OrgID:            orgID,
			StripeCustomerID: payload.Customer,
		})
	}

	periodEnd := time.Unix(sub.CurrentPeriodEnd, 0)
	if err := h.planAssigner.Assign(ctx, services.AssignInput{
		OrgID:                orgID,
		StripeSubscriptionID: sub.ID,
		StripePriceID:        sub.PriceID,
		StripeCustomerID:     payload.Customer,
		BillingStatus:        entities.BillingActive,
		CurrentPeriodEnd:     periodEnd,
		PaymentStatus:        "ok",
	}); err != nil {
		return err
	}

	// Trial conversion (best-effort). When the org has just paid for the
	// first time we re-resolve the orgID (it can be uuid.Nil on the very
	// first delivery before customer mapping existed) and mark the trial
	// plan rows as converted.
	if h.trialConverter != nil {
		convertOrgID := orgID
		if convertOrgID == uuid.Nil {
			if c, _ := h.customerRepo.FindByStripeCustomerID(ctx, payload.Customer); c != nil {
				convertOrgID = c.OrgID
			}
		}
		if convertOrgID != uuid.Nil {
			if err := h.trialConverter.Convert(ctx, convertOrgID); err != nil {
				// Log only — conversion bookkeeping should not fail the
				// webhook (Stripe would retry indefinitely otherwise).
				logger.Warn("trial converter failed on checkout.session.completed",
					zap.String("event_id", event.ID),
					zap.String("org_id", convertOrgID.String()),
					zap.Error(err),
				)
			}
		}
	}
	return nil
}

func (h *Handler) handleInvoicePaid(ctx context.Context, event services.StripeEvent) error {
	var payload invoicePayload
	if err := json.Unmarshal(event.RawData, &payload); err != nil {
		return err
	}

	sub, err := h.gateway.RetrieveSubscription(ctx, payload.Subscription)
	if err != nil {
		return err
	}

	orgID, err := h.resolveOrgByCustomer(ctx, payload.Customer)
	if err != nil {
		return err
	}

	periodEnd := time.Unix(sub.CurrentPeriodEnd, 0)
	return h.planAssigner.Assign(ctx, services.AssignInput{
		OrgID:                orgID,
		StripeSubscriptionID: sub.ID,
		StripePriceID:        sub.PriceID,
		StripeCustomerID:     payload.Customer,
		BillingStatus:        entities.BillingActive,
		CurrentPeriodEnd:     periodEnd,
		PaymentStatus:        "ok",
	})
}

func (h *Handler) handleSubscriptionUpdated(ctx context.Context, event services.StripeEvent) error {
	var payload subscriptionPayload
	if err := json.Unmarshal(event.RawData, &payload); err != nil {
		return err
	}

	orgID, err := h.resolveOrgByCustomer(ctx, payload.Customer)
	if err != nil {
		return err
	}

	var priceID string
	if len(payload.Items.Data) > 0 {
		priceID = payload.Items.Data[0].Price.ID
	}

	status, _ := entities.BillingStatusFromString(payload.Status)
	periodEnd := time.Unix(payload.resolveCurrentPeriodEnd(), 0)

	err = h.planAssigner.Assign(ctx, services.AssignInput{
		OrgID:                orgID,
		StripeSubscriptionID: payload.ID,
		StripePriceID:        priceID,
		StripeCustomerID:     payload.Customer,
		BillingStatus:        status,
		CurrentPeriodEnd:     periodEnd,
		PaymentStatus:        "ok",
	})
	if errors.Is(err, services.ErrPlanNotFound) {
		// FR5: unknown price_id — log a warning and acknowledge the event (200).
		// Do NOT propagate: Stripe would retry indefinitely causing a storm.
		logger.Warn("billing: subscription.updated with unknown price_id — ignoring",
			zap.String("price_id", priceID),
			zap.String("subscription_id", payload.ID),
		)
		return nil
	}
	return err
}

func (h *Handler) handleSubscriptionDeleted(ctx context.Context, event services.StripeEvent) error {
	var payload subscriptionPayload
	if err := json.Unmarshal(event.RawData, &payload); err != nil {
		return err
	}

	// Full cancellation → deactivate the org's plan with NO replacement. The org
	// ends up with no active plan (expired): read-only / gated until they
	// subscribe again. (Previously this downgraded to starter, which silently
	// kept the org on a paid-looking tier after they canceled.)
	return h.planAssigner.Deactivate(ctx, payload.Customer)
}

func (h *Handler) handlePaymentFailed(ctx context.Context, event services.StripeEvent) error {
	var payload invoicePayload
	if err := json.Unmarshal(event.RawData, &payload); err != nil {
		return err
	}

	sub, err := h.gateway.RetrieveSubscription(ctx, payload.Subscription)
	if err != nil {
		return err
	}

	orgID, err := h.resolveOrgByCustomer(ctx, payload.Customer)
	if err != nil {
		return err
	}

	periodEnd := time.Unix(sub.CurrentPeriodEnd, 0)
	return h.planAssigner.Assign(ctx, services.AssignInput{
		OrgID:                orgID,
		StripeSubscriptionID: sub.ID,
		StripePriceID:        sub.PriceID,
		StripeCustomerID:     payload.Customer,
		BillingStatus:        entities.BillingPastDue,
		CurrentPeriodEnd:     periodEnd,
		PaymentStatus:        "grace_period",
	})
}

// handlePricingChanged caches a Stripe Price's amount into public.plans so the
// CMS can render the real charged price. It is deliberately tolerant: any event
// it cannot map (no plan_code, non-recurring price, unknown plan) is logged and
// acknowledged so Stripe does not retry.
func (h *Handler) handlePricingChanged(ctx context.Context, event services.StripeEvent) error {
	if h.planRepo == nil {
		return nil // pricing-sync not wired
	}

	var payload pricePayload
	if err := json.Unmarshal(event.RawData, &payload); err != nil {
		return err
	}

	if payload.Recurring == nil || payload.Recurring.Interval == "" {
		return nil
	}

	planCode := payload.Metadata.PlanCode
	if planCode == "" {
		logger.Warn("billing: pricing event without metadata.plan_code — ignoring",
			zap.String("event_id", event.ID),
			zap.String("price_id", payload.ID),
		)
		return nil
	}

	var cycle string
	switch payload.Recurring.Interval {
	case "month":
		cycle = "monthly"
	case "year":
		cycle = "yearly"
	default:
		logger.Warn("billing: pricing event with unsupported interval — ignoring",
			zap.String("interval", payload.Recurring.Interval),
			zap.String("price_id", payload.ID),
		)
		return nil
	}

	err := h.planRepo.UpsertStripePricing(ctx, planCode, cycle, payload.ID, payload.UnitAmount, payload.Currency)
	if errors.Is(err, repositories.ErrPlanNotFound) {
		logger.Warn("billing: pricing event for unknown plan_code — ignoring",
			zap.String("plan_code", planCode),
			zap.String("price_id", payload.ID),
		)
		return nil
	}
	return err
}

// resolveOrgByCustomer looks up the orgID associated with a Stripe customer ID.
// Returns an error if the customer is not found (org has never completed checkout).
func (h *Handler) resolveOrgByCustomer(ctx context.Context, customerID string) (uuid.UUID, error) {
	customer, err := h.customerRepo.FindByStripeCustomerID(ctx, customerID)
	if err != nil {
		return uuid.Nil, err
	}
	if customer == nil {
		// Customer not yet persisted locally — this can happen on checkout.session.completed
		// before we've stored the mapping. We fall back to uuid.Nil so the caller can
		// optionally create the mapping. The PlanAssigner's UPSERT handles this via customerID.
		return uuid.Nil, nil
	}
	return customer.OrgID, nil
}
