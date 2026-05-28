// Package stripe provides the concrete implementation of the billing StripeGateway interface.
// It wraps github.com/stripe/stripe-go/v79 and maps Stripe SDK types to domain types.
package stripe

import (
	"context"
	"fmt"
	"time"

	stripe "github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/coupon"
	"github.com/stripe/stripe-go/v79/customer"
	"github.com/stripe/stripe-go/v79/customerbalancetransaction"
	"github.com/stripe/stripe-go/v79/invoice"
	"github.com/stripe/stripe-go/v79/invoiceitem"
	"github.com/stripe/stripe-go/v79/price"
	"github.com/stripe/stripe-go/v79/promotioncode"
	"github.com/stripe/stripe-go/v79/subscription"
	"github.com/stripe/stripe-go/v79/subscriptionschedule"
	"github.com/stripe/stripe-go/v79/webhook"

	billingservices "github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"

	billingportalclient "github.com/stripe/stripe-go/v79/billingportal/session"
)

// Compile-time interface check.
var _ billingservices.StripeGateway = (*Gateway)(nil)

// Gateway is the concrete Stripe implementation of billing.StripeGateway.
// It is constructed with the Stripe API key and webhook secret from config.
// The stripe-go SDK is initialised in the constructor via stripe.Key global.
type Gateway struct {
	webhookSecret string
}

// NewGateway creates a Gateway and sets the Stripe API key on the global SDK client.
// apiKey must be a live or test secret key (sk_live_* or sk_test_*).
// webhookSecret must be the signing secret from the Stripe webhook dashboard (whsec_*).
func NewGateway(apiKey, webhookSecret string) *Gateway {
	stripe.Key = apiKey
	return &Gateway{webhookSecret: webhookSecret}
}

// EnsureCustomer returns the existing Stripe customer ID for the org, or creates one.
// It searches by email first to avoid duplicate Stripe customers across retries.
func (g *Gateway) EnsureCustomer(_ context.Context, orgID, email, name string) (string, error) {
	// Look up by email — Stripe allows filtering customers by exact email.
	params := &stripe.CustomerListParams{}
	params.Email = stripe.String(email)
	params.Limit = stripe.Int64(1)

	iter := customer.List(params)
	for iter.Next() {
		return iter.Customer().ID, nil
	}
	if err := iter.Err(); err != nil {
		return "", fmt.Errorf("billing: stripe customer lookup: %w", err)
	}

	// No existing customer — create one.
	created, err := customer.New(&stripe.CustomerParams{
		Email:    stripe.String(email),
		Name:     stripe.String(name),
		Metadata: map[string]string{"org_id": orgID},
	})
	if err != nil {
		return "", fmt.Errorf("billing: stripe customer create: %w", err)
	}

	return created.ID, nil
}

// CreateCheckoutSession creates a hosted Stripe Checkout session for a subscription.
func (g *Gateway) CreateCheckoutSession(_ context.Context, in billingservices.CheckoutInput) (string, error) {
	qty := stripe.Int64(1)
	mode := string(stripe.CheckoutSessionModeSubscription)

	params := &stripe.CheckoutSessionParams{
		Customer:   stripe.String(in.CustomerID),
		SuccessURL: stripe.String(in.SuccessURL),
		CancelURL:  stripe.String(in.CancelURL),
		Mode:       stripe.String(mode),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(in.PriceID),
				Quantity: qty,
			},
		},
	}

	// Discounts and AllowPromotionCodes are mutually exclusive in Stripe.
	// Auto-apply link path → pre-apply the promotion code. Otherwise expose
	// the hosted promo-code field so customers can type one manually.
	if in.PromotionCodeID != "" {
		params.Discounts = []*stripe.CheckoutSessionDiscountParams{
			{PromotionCode: stripe.String(in.PromotionCodeID)},
		}
	} else {
		params.AllowPromotionCodes = stripe.Bool(true)
	}

	s, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("billing: stripe checkout session: %w", err)
	}

	return s.URL, nil
}

// CreatePortalSession creates a Stripe Customer Portal session and returns its URL.
func (g *Gateway) CreatePortalSession(_ context.Context, customerID, returnURL string) (string, error) {
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(returnURL),
	}

	s, err := billingportalclient.New(params)
	if err != nil {
		return "", fmt.Errorf("billing: stripe portal session: %w", err)
	}

	return s.URL, nil
}

// ConstructEvent validates the Stripe-Signature header and parses the raw webhook payload.
// Returns ErrInvalidSignature (via the caller) on any validation failure.
func (g *Gateway) ConstructEvent(payload []byte, signature, secret string) (billingservices.StripeEvent, error) {
	event, err := webhook.ConstructEventWithOptions(payload, signature, secret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		return billingservices.StripeEvent{}, err
	}

	raw := []byte{}
	if event.Data != nil {
		raw = []byte(event.Data.Raw)
	}

	return billingservices.StripeEvent{
		ID:      event.ID,
		Type:    string(event.Type),
		RawData: raw,
	}, nil
}

// RetrieveSubscription fetches the current subscription state from Stripe.
func (g *Gateway) RetrieveSubscription(_ context.Context, subID string) (billingservices.StripeSubscription, error) {
	sub, err := subscription.Get(subID, nil)
	if err != nil {
		return billingservices.StripeSubscription{}, fmt.Errorf("billing: stripe retrieve subscription: %w", err)
	}

	var priceID string
	if sub.Items != nil && len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
		priceID = sub.Items.Data[0].Price.ID
	}

	var customerID string
	if sub.Customer != nil {
		customerID = sub.Customer.ID
	}

	out := billingservices.StripeSubscription{
		ID:                sub.ID,
		Status:            string(sub.Status),
		CurrentPeriodEnd:  sub.CurrentPeriodEnd,
		CustomerID:        customerID,
		PriceID:           priceID,
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
		CancelAt:          sub.CancelAt,
	}
	// Surface any active discount (coupon) so the UI can show a gift banner.
	if sub.Discount != nil && sub.Discount.Coupon != nil {
		c := sub.Discount.Coupon
		out.DiscountAmountOffCents = c.AmountOff
		out.DiscountPercentOff = int64(c.PercentOff)
		if c.Metadata != nil {
			out.DiscountPurpose = c.Metadata["purpose"]
		}
	}
	// Surface plan-gift state from subscription metadata (set by a gift
	// schedule). Phase 2 clears gift_active, so this is true only while the
	// free gifted-plan phase is running.
	if sub.Metadata != nil && sub.Metadata["gift_active"] == "1" {
		out.GiftPlanActive = true
		out.GiftPlanCode = sub.Metadata["gift_plan"]
		out.GiftRevertPlanCode = sub.Metadata["gift_revert_plan"]
	}
	return out, nil
}

// GiftPlanSchedule converts subID into a 2-phase schedule: a free gifted-plan
// phase (1 cycle, trial) then a revert to the current plan ongoing.
func (g *Gateway) GiftPlanSchedule(_ context.Context, subID, giftPriceID, currentPriceID, giftCode, currentCode string) (int64, error) {
	sched, err := subscriptionschedule.New(&stripe.SubscriptionScheduleParams{
		FromSubscription: stripe.String(subID),
	})
	if err != nil {
		return 0, fmt.Errorf("billing: stripe create gift schedule: %w", err)
	}

	updated, err := subscriptionschedule.Update(sched.ID, &stripe.SubscriptionScheduleParams{
		EndBehavior:       stripe.String("release"),
		ProrationBehavior: stripe.String("none"),
		Phases: []*stripe.SubscriptionSchedulePhaseParams{
			{
				StartDateNow: stripe.Bool(true),
				Items: []*stripe.SubscriptionSchedulePhaseItemParams{
					{Price: stripe.String(giftPriceID)},
				},
				Iterations: stripe.Int64(1),
				Trial:      stripe.Bool(true),
				Metadata: map[string]string{
					"gift_active":      "1",
					"gift_plan":        giftCode,
					"gift_revert_plan": currentCode,
					"purpose":          "gift_plan_month",
				},
			},
			{
				Items: []*stripe.SubscriptionSchedulePhaseItemParams{
					{Price: stripe.String(currentPriceID)},
				},
				Metadata: map[string]string{
					"gift_active": "",
				},
			},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("billing: stripe update gift schedule: %w", err)
	}

	if len(updated.Phases) > 0 {
		return updated.Phases[0].EndDate, nil
	}
	return 0, nil
}

// ListSubscriptions returns all subscriptions Stripe holds for the customer.
// Caller filters by Status. Items[0].Price.ID is exposed as PriceID for parity
// with RetrieveSubscription. Pagination uses stripe-go's auto-iteration.
func (g *Gateway) ListSubscriptions(_ context.Context, customerID string) ([]billingservices.StripeSubscription, error) {
	params := &stripe.SubscriptionListParams{
		Customer: stripe.String(customerID),
	}
	// Stripe defaults to status=incomplete-active mix; we want all so the caller can decide.
	params.Status = stripe.String("all")
	params.Filters.AddFilter("limit", "", "100")

	iter := subscription.List(params)
	var out []billingservices.StripeSubscription
	for iter.Next() {
		sub := iter.Subscription()
		var priceID string
		if sub.Items != nil && len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
			priceID = sub.Items.Data[0].Price.ID
		}
		var custID string
		if sub.Customer != nil {
			custID = sub.Customer.ID
		}
		out = append(out, billingservices.StripeSubscription{
			ID:               sub.ID,
			Status:           string(sub.Status),
			CurrentPeriodEnd: sub.CurrentPeriodEnd,
			CustomerID:       custID,
			PriceID:          priceID,
		})
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("billing: stripe list subscriptions: %w", err)
	}
	return out, nil
}

// UpdateSubscriptionItem swaps the first item's price on subID to newPriceID
// and lets Stripe handle the proration via the supplied prorationBehavior.
// Returns the refreshed subscription so callers can persist new period_end.
func (g *Gateway) UpdateSubscriptionItem(_ context.Context, subID, newPriceID, prorationBehavior string) (billingservices.StripeSubscription, error) {
	current, err := subscription.Get(subID, nil)
	if err != nil {
		return billingservices.StripeSubscription{}, fmt.Errorf("billing: stripe get sub: %w", err)
	}
	if current.Items == nil || len(current.Items.Data) == 0 {
		return billingservices.StripeSubscription{}, fmt.Errorf("billing: subscription %s has no items", subID)
	}
	itemID := current.Items.Data[0].ID

	if prorationBehavior == "" {
		prorationBehavior = "create_prorations"
	}

	params := &stripe.SubscriptionParams{
		ProrationBehavior: stripe.String(prorationBehavior),
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(itemID),
				Price: stripe.String(newPriceID),
			},
		},
	}
	updated, err := subscription.Update(subID, params)
	if err != nil {
		return billingservices.StripeSubscription{}, fmt.Errorf("billing: stripe subscription update: %w", err)
	}

	var priceID string
	if updated.Items != nil && len(updated.Items.Data) > 0 && updated.Items.Data[0].Price != nil {
		priceID = updated.Items.Data[0].Price.ID
	}
	var customerID string
	if updated.Customer != nil {
		customerID = updated.Customer.ID
	}
	return billingservices.StripeSubscription{
		ID:               updated.ID,
		Status:           string(updated.Status),
		CurrentPeriodEnd: updated.CurrentPeriodEnd,
		CustomerID:       customerID,
		PriceID:          priceID,
	}, nil
}

// PreviewProration returns the prorated amount that would be charged today if
// the subscription's price changed to newPriceID right now. Uses the upcoming
// invoice preview endpoint with subscription_proration_date=now so Stripe
// computes the credit for the unused portion of the current plan and only the
// proration line items are returned (not the next full cycle invoice).
func (g *Gateway) PreviewProration(_ context.Context, subID, newPriceID string) (int64, string, error) {
	current, err := subscription.Get(subID, nil)
	if err != nil {
		return 0, "", fmt.Errorf("billing: stripe get sub: %w", err)
	}
	if current.Items == nil || len(current.Items.Data) == 0 {
		return 0, "", fmt.Errorf("billing: subscription %s has no items", subID)
	}
	itemID := current.Items.Data[0].ID
	customerID := ""
	if current.Customer != nil {
		customerID = current.Customer.ID
	}

	prorationTimestamp := time.Now().Unix()
	params := &stripe.InvoiceUpcomingParams{
		Customer:     stripe.String(customerID),
		Subscription: stripe.String(subID),
		SubscriptionItems: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(itemID),
				Price: stripe.String(newPriceID),
			},
		},
		SubscriptionProrationBehavior: stripe.String("create_prorations"),
		SubscriptionProrationDate:     stripe.Int64(prorationTimestamp),
	}
	preview, err := invoice.Upcoming(params)
	if err != nil {
		return 0, "", fmt.Errorf("billing: stripe invoice upcoming: %w", err)
	}

	// Sum only the proration line items so we report the delta charged today,
	// not the next full cycle invoice. Proration entries are flagged by the
	// `proration` field on each line. Falling back to AmountDue is misleading
	// when Stripe returns the upcoming-cycle invoice (full new plan price).
	var prorationCents int64
	if preview.Lines != nil {
		for _, line := range preview.Lines.Data {
			if line == nil || !line.Proration {
				continue
			}
			prorationCents += line.Amount
		}
	}
	// When Stripe found no proration lines (e.g. switching to a price with
	// the same recurring amount), fall through to AmountDue so we never show
	// a $0 charge on what could still be a real upgrade.
	if prorationCents == 0 {
		prorationCents = preview.AmountDue
	}
	if prorationCents < 0 {
		prorationCents = 0
	}
	return prorationCents, string(preview.Currency), nil
}

// CreateRefundCreditInvoiceItem attaches a negative-amount InvoiceItem to
// customerID so the next invoice is reduced by amountCents. amountCents must
// be positive; the SDK requires the negative-sign explicit on Amount.
func (g *Gateway) CreateRefundCreditInvoiceItem(_ context.Context, customerID string, amountCents int64, currency, description string) error {
	if amountCents <= 0 {
		return nil
	}
	if currency == "" {
		currency = "usd"
	}
	params := &stripe.InvoiceItemParams{
		Customer:    stripe.String(customerID),
		Amount:      stripe.Int64(-amountCents),
		Currency:    stripe.String(currency),
		Description: stripe.String(description),
	}
	if _, err := invoiceitem.New(params); err != nil {
		return fmt.Errorf("billing: stripe create refund credit invoice item: %w", err)
	}
	return nil
}

// UpdateSubscriptionAnchorNow swaps the subscription's price to newPriceID,
// resets the billing cycle to start today, and turns proration OFF — the
// caller is expected to have applied any custom credit (e.g. usage-based
// refund) via a negative InvoiceItem before calling this.
func (g *Gateway) UpdateSubscriptionAnchorNow(_ context.Context, subID, newPriceID string) (billingservices.StripeSubscription, error) {
	current, err := subscription.Get(subID, nil)
	if err != nil {
		return billingservices.StripeSubscription{}, fmt.Errorf("billing: stripe get sub: %w", err)
	}
	if current.Items == nil || len(current.Items.Data) == 0 {
		return billingservices.StripeSubscription{}, fmt.Errorf("billing: subscription %s has no items", subID)
	}
	itemID := current.Items.Data[0].ID

	params := &stripe.SubscriptionParams{
		ProrationBehavior:     stripe.String("none"),
		BillingCycleAnchorNow: stripe.Bool(true),
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(itemID),
				Price: stripe.String(newPriceID),
			},
		},
	}
	updated, err := subscription.Update(subID, params)
	if err != nil {
		return billingservices.StripeSubscription{}, fmt.Errorf("billing: stripe subscription update (anchor=now): %w", err)
	}

	var priceID string
	if updated.Items != nil && len(updated.Items.Data) > 0 && updated.Items.Data[0].Price != nil {
		priceID = updated.Items.Data[0].Price.ID
	}
	var customerID string
	if updated.Customer != nil {
		customerID = updated.Customer.ID
	}
	return billingservices.StripeSubscription{
		ID:               updated.ID,
		Status:           string(updated.Status),
		CurrentPeriodEnd: updated.CurrentPeriodEnd,
		CustomerID:       customerID,
		PriceID:          priceID,
	}, nil
}

// RetrievePriceAmount returns the unit_amount and currency of a Stripe Price.
func (g *Gateway) RetrievePriceAmount(_ context.Context, priceID string) (int64, string, error) {
	p, err := price.Get(priceID, nil)
	if err != nil {
		return 0, "", fmt.Errorf("billing: stripe get price: %w", err)
	}
	return p.UnitAmount, string(p.Currency), nil
}

// RetrieveCustomerBalance returns the customer's account balance and currency.
func (g *Gateway) RetrieveCustomerBalance(_ context.Context, customerID string) (int64, string, error) {
	c, err := customer.Get(customerID, nil)
	if err != nil {
		return 0, "", fmt.Errorf("billing: stripe retrieve customer: %w", err)
	}
	return c.Balance, string(c.Currency), nil
}

// CreditCustomerBalance grants the customer a balance credit. amountCents is
// positive; Stripe records credit as a NEGATIVE balance transaction.
func (g *Gateway) CreditCustomerBalance(_ context.Context, customerID string, amountCents int64, currency, description string) error {
	if amountCents <= 0 {
		return fmt.Errorf("billing: credit amount must be positive, got %d", amountCents)
	}
	if currency == "" {
		currency = "usd"
	}
	params := &stripe.CustomerBalanceTransactionParams{
		Customer: stripe.String(customerID),
		Amount:   stripe.Int64(-amountCents), // negative = credit to customer
		Currency: stripe.String(currency),
	}
	if description != "" {
		params.Description = stripe.String(description)
	}
	if _, err := customerbalancetransaction.New(params); err != nil {
		return fmt.Errorf("billing: stripe credit customer balance: %w", err)
	}
	return nil
}

// ── Coupons / promotion codes ───────────────────────────────────────────────

// CreateCoupon creates a once-applied amount_off coupon and returns its ID.
func (g *Gateway) CreateCoupon(_ context.Context, amountOffCents int64, currency string, meta billingservices.CouponMetadata) (string, error) {
	if currency == "" {
		currency = "usd"
	}
	params := &stripe.CouponParams{
		AmountOff: stripe.Int64(amountOffCents),
		Currency:  stripe.String(currency),
		Duration:  stripe.String(string(stripe.CouponDurationOnce)),
	}
	params.AddMetadata("plan_code", meta.PlanCode)
	params.AddMetadata("billing_cycle", meta.BillingCycle)
	params.AddMetadata("purpose", meta.Purpose)

	c, err := coupon.New(params)
	if err != nil {
		return "", fmt.Errorf("billing: stripe create coupon: %w", err)
	}
	return c.ID, nil
}

// ApplyCouponToSubscription attaches couponID to an existing subscription so
// the discount lands on the next invoice.
func (g *Gateway) ApplyCouponToSubscription(_ context.Context, subID, couponID string) error {
	_, err := subscription.Update(subID, &stripe.SubscriptionParams{
		Coupon: stripe.String(couponID),
	})
	if err != nil {
		return fmt.Errorf("billing: stripe apply coupon to subscription: %w", err)
	}
	return nil
}

// CreatePromotionCode wraps a coupon in a customer-facing code.
func (g *Gateway) CreatePromotionCode(_ context.Context, couponID, code string, maxRedemptions, expiresAt int64) (billingservices.PromotionCode, error) {
	params := &stripe.PromotionCodeParams{
		Coupon: stripe.String(couponID),
	}
	if code != "" {
		params.Code = stripe.String(code)
	}
	if maxRedemptions > 0 {
		params.MaxRedemptions = stripe.Int64(maxRedemptions)
	}
	if expiresAt > 0 {
		params.ExpiresAt = stripe.Int64(expiresAt)
	}

	pc, err := promotioncode.New(params)
	if err != nil {
		return billingservices.PromotionCode{}, fmt.Errorf("billing: stripe create promotion code: %w", err)
	}
	return toDomainPromotionCode(pc), nil
}

// ListPromotionCodes returns all promotion codes (active + inactive).
func (g *Gateway) ListPromotionCodes(_ context.Context) ([]billingservices.PromotionCode, error) {
	params := &stripe.PromotionCodeListParams{}
	params.Filters.AddFilter("limit", "", "100")

	iter := promotioncode.List(params)
	var out []billingservices.PromotionCode
	for iter.Next() {
		out = append(out, toDomainPromotionCode(iter.PromotionCode()))
	}
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("billing: stripe list promotion codes: %w", err)
	}
	return out, nil
}

// DeactivatePromotionCode flips a promotion code inactive.
func (g *Gateway) DeactivatePromotionCode(_ context.Context, promotionCodeID string) error {
	_, err := promotioncode.Update(promotionCodeID, &stripe.PromotionCodeParams{
		Active: stripe.Bool(false),
	})
	if err != nil {
		return fmt.Errorf("billing: stripe deactivate promotion code: %w", err)
	}
	return nil
}

// FindPromotionCodeByCode resolves a human code to its (active) promotion code.
func (g *Gateway) FindPromotionCodeByCode(_ context.Context, code string) (billingservices.PromotionCode, error) {
	params := &stripe.PromotionCodeListParams{
		Code:   stripe.String(code),
		Active: stripe.Bool(true),
	}
	params.Filters.AddFilter("limit", "", "1")

	iter := promotioncode.List(params)
	for iter.Next() {
		return toDomainPromotionCode(iter.PromotionCode()), nil
	}
	if err := iter.Err(); err != nil {
		return billingservices.PromotionCode{}, fmt.Errorf("billing: stripe find promotion code: %w", err)
	}
	return billingservices.PromotionCode{}, fmt.Errorf("billing: promotion code %q not found or inactive", code)
}

// ── Cancellation ────────────────────────────────────────────────────────────

// CancelSubscriptionAtPeriodEnd flags the subscription to end at period close.
func (g *Gateway) CancelSubscriptionAtPeriodEnd(_ context.Context, subID string) (billingservices.StripeSubscription, error) {
	updated, err := subscription.Update(subID, &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	})
	if err != nil {
		return billingservices.StripeSubscription{}, fmt.Errorf("billing: stripe cancel subscription at period end: %w", err)
	}
	return toDomainSubscription(updated), nil
}

// ResumeSubscription clears a pending period-end cancellation.
func (g *Gateway) ResumeSubscription(_ context.Context, subID string) (billingservices.StripeSubscription, error) {
	updated, err := subscription.Update(subID, &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(false),
	})
	if err != nil {
		return billingservices.StripeSubscription{}, fmt.Errorf("billing: stripe resume subscription: %w", err)
	}
	return toDomainSubscription(updated), nil
}

// RemoveSubscriptionDiscount deletes the active discount from a subscription.
func (g *Gateway) RemoveSubscriptionDiscount(_ context.Context, subID string) error {
	if _, err := subscription.DeleteDiscount(subID, nil); err != nil {
		return fmt.Errorf("billing: stripe delete subscription discount: %w", err)
	}
	return nil
}

// CreateGiftSubscription creates a free one-month trial subscription on
// giftPriceID for a customer with no active subscription. With no payment
// method, the trial auto-cancels at the end (missing_payment_method=cancel).
func (g *Gateway) CreateGiftSubscription(_ context.Context, customerID, giftPriceID, giftCode string) (billingservices.StripeSubscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{Price: stripe.String(giftPriceID)},
		},
		TrialPeriodDays: stripe.Int64(30),
		TrialSettings: &stripe.SubscriptionTrialSettingsParams{
			EndBehavior: &stripe.SubscriptionTrialSettingsEndBehaviorParams{
				MissingPaymentMethod: stripe.String("cancel"),
			},
		},
		Metadata: map[string]string{
			"gift_active":      "1",
			"gift_plan":        giftCode,
			"gift_revert_plan": "",
			"purpose":          "gift_new_user",
		},
	}
	sub, err := subscription.New(params)
	if err != nil {
		return billingservices.StripeSubscription{}, fmt.Errorf("billing: stripe create gift subscription: %w", err)
	}
	return toDomainSubscription(sub), nil
}

// toDomainSubscription maps a stripe.Subscription to the domain shape.
func toDomainSubscription(sub *stripe.Subscription) billingservices.StripeSubscription {
	var priceID string
	if sub.Items != nil && len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
		priceID = sub.Items.Data[0].Price.ID
	}
	var customerID string
	if sub.Customer != nil {
		customerID = sub.Customer.ID
	}
	out := billingservices.StripeSubscription{
		ID:                sub.ID,
		Status:            string(sub.Status),
		CurrentPeriodEnd:  sub.CurrentPeriodEnd,
		CustomerID:        customerID,
		PriceID:           priceID,
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
		CancelAt:          sub.CancelAt,
	}
	if sub.Metadata != nil && sub.Metadata["gift_active"] == "1" {
		out.GiftPlanActive = true
		out.GiftPlanCode = sub.Metadata["gift_plan"]
		out.GiftRevertPlanCode = sub.Metadata["gift_revert_plan"]
	}
	return out
}

func toDomainPromotionCode(pc *stripe.PromotionCode) billingservices.PromotionCode {
	out := billingservices.PromotionCode{
		ID:             pc.ID,
		Code:           pc.Code,
		Active:         pc.Active,
		MaxRedemptions: pc.MaxRedemptions,
		TimesRedeemed:  pc.TimesRedeemed,
		ExpiresAt:      pc.ExpiresAt,
	}
	if pc.Coupon != nil {
		out.CouponID = pc.Coupon.ID
		out.AmountOffCents = pc.Coupon.AmountOff
		out.Currency = string(pc.Coupon.Currency)
		if pc.Coupon.Metadata != nil {
			out.PlanCode = pc.Coupon.Metadata["plan_code"]
			out.BillingCycle = pc.Coupon.Metadata["billing_cycle"]
		}
	}
	return out
}
