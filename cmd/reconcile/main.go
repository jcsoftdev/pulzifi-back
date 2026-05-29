// Command reconcile materialises Stripe's view of a customer's subscriptions
// into public.organization_plans. Intended for one-shot ops use: when a Stripe
// payment landed before the local org/customer mapping existed (orphan webhook
// scenario) or when a deferred event needs to be replayed manually.
//
// Usage:
//
//	go run ./cmd/reconcile -org=<org_uuid> -cust=<stripe_customer_id>
//
// The binary connects to the same database as the API server via shared/config
// and shared/database, exits 0 on success and a non-zero code on any error.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/uuid"
	billingwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/billing"
	reconcilesubscription "github.com/jcsoftdev/pulzifi-back/modules/billing/application/reconcile_subscription"
	billingpostgres "github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/persistence/postgres"
	billingstripe "github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/stripe"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
	"github.com/jcsoftdev/pulzifi-back/shared/database"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	os.Exit(run())
}

// run executes the reconcile flow and returns a process exit code. It is kept
// separate from main so that deferred cleanup (e.g. db.Close) runs before the
// process exits — os.Exit in main would skip deferred calls.
func run() int {
	orgFlag := flag.String("org", "", "Organization UUID (public.organizations.id)")
	custFlag := flag.String("cust", "", "Stripe customer ID (cus_...)")
	envFlag := flag.String("env", "", "Optional environment suffix; loads .env.<env> (e.g. -env=production loads .env.production). Overrides the default .env.")
	flag.Parse()

	if *orgFlag == "" || *custFlag == "" {
		fmt.Fprintln(os.Stderr, "usage: reconcile -org=<uuid> -cust=<cus_...> [-env=production]")
		return 2
	}

	orgID, err := uuid.Parse(*orgFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid -org: %v\n", err)
		return 2
	}

	// Load the requested env file BEFORE config.Load() so its values win.
	// godotenv.Overload also overrides shell vars, which is what we want when
	// the operator explicitly points at .env.production.
	if *envFlag != "" {
		envPath := ".env." + *envFlag
		if err := godotenv.Overload(envPath); err != nil {
			fmt.Fprintf(os.Stderr, "load %s: %v\n", envPath, err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "reconcile: loaded %s\n", envPath)
	}

	cfg := config.Load()
	if !cfg.BillingEnabled {
		fmt.Fprintln(os.Stderr, "BILLING_ENABLED is false — refusing to run reconcile")
		return 1
	}
	if cfg.StripeSecretKey == "" {
		fmt.Fprintln(os.Stderr, "STRIPE_SECRET_KEY is required")
		return 1
	}

	logger.Info("reconcile: starting",
		zap.String("org_id", orgID.String()),
		zap.String("stripe_customer_id", *custFlag),
	)

	db, err := database.Connect(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db connect: %v\n", err)
		return 1
	}
	defer func() { _ = db.Close() }()

	gateway := billingstripe.NewGateway(cfg.StripeSecretKey, cfg.StripeWebhookSecret)
	planAssigner := billingwiring.NewPlanAssigner(db)
	webhookRepo := billingpostgres.NewWebhookEventPostgresRepository(db)

	handler := reconcilesubscription.NewHandler(gateway, planAssigner, webhookRepo)

	ctx := context.Background()
	res, err := handler.Handle(ctx, reconcilesubscription.Input{
		OrgID:            orgID,
		StripeCustomerID: *custFlag,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "reconcile failed: %v\n", err)
		return 1
	}

	logger.Info("reconcile: done",
		zap.Int("subscriptions_found", res.SubscriptionsFound),
		zap.Int("subscriptions_applied", res.SubscriptionsApplied),
		zap.Int("deferred_events_processed", res.DeferredEventsProcessed),
	)
	fmt.Printf("OK — found=%d applied=%d deferred_processed=%d\n",
		res.SubscriptionsFound, res.SubscriptionsApplied, res.DeferredEventsProcessed)
	return 0
}
