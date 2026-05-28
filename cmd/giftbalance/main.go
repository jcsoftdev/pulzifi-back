// Command giftbalance migrates a legacy gift from an amount_off coupon into a
// Stripe customer BALANCE credit. Older gifts were applied as a once-only
// coupon on the subscription, which collapses onto a single invoice and
// discards leftover value (e.g. a $62 gift used on a $35 upgrade loses $27).
// Balance credit auto-applies across invoices and preserves leftover, and is
// what update_subscription proration subtracts.
//
// For the target subscription it:
//  1. Reads the active discount (must have purpose=gift_one_month).
//  2. Deletes the discount from the subscription.
//  3. Credits the customer's balance with the coupon's amount_off.
//
// Usage:
//   go run ./cmd/giftbalance -sub=<sub_...> [-env=production]
//   go run ./cmd/giftbalance -sub=<sub_...> -force   # migrate any discount, not just gifts
//
// Idempotency: if the subscription has no discount, it exits 0 with a no-op.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	billingstripe "github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/stripe"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
)

func main() {
	subFlag := flag.String("sub", "", "Stripe subscription ID (sub_...) carrying the gift coupon")
	envFlag := flag.String("env", "", "Optional environment suffix; loads .env.<env> (e.g. -env=production).")
	forceFlag := flag.Bool("force", false, "Migrate any discount, not only purpose=gift_one_month")
	flag.Parse()

	if *subFlag == "" {
		fmt.Fprintln(os.Stderr, "usage: giftbalance -sub=<sub_...> [-env=production] [-force]")
		os.Exit(2)
	}

	if *envFlag != "" {
		envPath := ".env." + *envFlag
		if err := godotenv.Overload(envPath); err != nil {
			fmt.Fprintf(os.Stderr, "load %s: %v\n", envPath, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "giftbalance: loaded %s\n", envPath)
	} else if err := godotenv.Overload(".env"); err == nil {
		fmt.Fprintln(os.Stderr, "giftbalance: loaded .env (overriding shell env)")
	}

	cfg := config.Load()
	if cfg.StripeSecretKey == "" {
		fmt.Fprintln(os.Stderr, "STRIPE_SECRET_KEY is required")
		os.Exit(1)
	}

	gateway := billingstripe.NewGateway(cfg.StripeSecretKey, cfg.StripeWebhookSecret)
	ctx := context.Background()

	sub, err := gateway.RetrieveSubscription(ctx, *subFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "retrieve subscription: %v\n", err)
		os.Exit(1)
	}

	if sub.DiscountAmountOffCents <= 0 {
		fmt.Printf("no amount_off discount on %s — nothing to migrate\n", *subFlag)
		return
	}
	if !*forceFlag && sub.DiscountPurpose != "gift_one_month" {
		fmt.Fprintf(os.Stderr,
			"discount purpose is %q, not gift_one_month — refusing (use -force to override)\n",
			sub.DiscountPurpose)
		os.Exit(1)
	}
	if sub.CustomerID == "" {
		fmt.Fprintln(os.Stderr, "subscription has no customer — cannot credit balance")
		os.Exit(1)
	}

	amount := sub.DiscountAmountOffCents
	desc := "Migrated gift coupon → balance credit"

	if err := gateway.RemoveSubscriptionDiscount(ctx, *subFlag); err != nil {
		fmt.Fprintf(os.Stderr, "remove discount: %v\n", err)
		os.Exit(1)
	}
	if err := gateway.CreditCustomerBalance(ctx, sub.CustomerID, amount, "", desc); err != nil {
		// Discount already removed — surface clearly so the operator can
		// re-credit manually rather than silently losing the gift.
		fmt.Fprintf(os.Stderr,
			"CRITICAL: discount removed but balance credit FAILED for customer %s amount %d: %v\n",
			sub.CustomerID, amount, err)
		os.Exit(1)
	}

	fmt.Printf("OK — migrated gift on %s: removed discount, credited %d cents to %s\n",
		*subFlag, amount, sub.CustomerID)
}
