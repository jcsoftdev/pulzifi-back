//go:build integration

package billingwiring_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
	billingwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/billing"
)

// openPlanAssignerTestDB returns a connected *sql.DB or calls t.Skip when
// DATABASE_URL (or DB_HOST) is not set.
func openPlanAssignerTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := planAssignerTestDSN()
	if dsn == "" {
		t.Skip("DATABASE_URL (or DB_HOST) not set; skipping billing wiring integration test")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("db.Ping: %v", err)
	}
	return db
}

func planAssignerTestDSN() string {
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		return dsn
	}
	host := os.Getenv("DB_HOST")
	if host == "" {
		return ""
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		host, port, os.Getenv("DB_NAME"),
	)
}

// seedTestOrg inserts a minimal user + organization row and registers cleanup.
// Returns the orgID.
func seedTestOrg(t *testing.T, db *sql.DB, ctx context.Context) (orgID uuid.UUID) {
	t.Helper()

	userID := uuid.New()
	orgID = uuid.New()
	suffix := orgID.String()[:8]

	t.Cleanup(func() {
		// Delete organization_plans first (FK), then org, then user.
		_, _ = db.ExecContext(ctx, `DELETE FROM public.organization_plans WHERE organization_id = $1`, orgID)
		_, _ = db.ExecContext(ctx, `DELETE FROM public.organizations WHERE id = $1`, orgID)
		_, _ = db.ExecContext(ctx, `DELETE FROM public.users WHERE id = $1`, userID)
	})

	_, err := db.ExecContext(ctx, `
		INSERT INTO public.users (id, email, password_hash, status)
		VALUES ($1, $2, 'testhash', 'approved')
		ON CONFLICT DO NOTHING`,
		userID, fmt.Sprintf("billingtest-%s@example.com", userID),
	)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO public.organizations (id, name, subdomain, schema_name, owner_user_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT DO NOTHING`,
		orgID,
		"BillingTest Org "+suffix,
		"billingtest-"+suffix,
		"billingtest_"+suffix,
		userID,
	)
	if err != nil {
		t.Fatalf("seed org: %v", err)
	}

	return orgID
}

// seedStarterPlan ensures a starter plan row exists and returns its ID.
// If a starter plan already exists, its ID is returned directly.
// The plan is NOT deleted in cleanup — it may already exist in prod fixtures.
func seedStarterPlan(t *testing.T, db *sql.DB, ctx context.Context) uuid.UUID {
	t.Helper()

	// First try to find an existing starter plan.
	var existing string
	err := db.QueryRowContext(ctx,
		`SELECT id::text FROM public.plans WHERE code = 'starter' AND is_active = TRUE LIMIT 1`,
	).Scan(&existing)
	if err == nil {
		id, parseErr := uuid.Parse(existing)
		if parseErr == nil {
			return id
		}
	}

	// No starter plan — insert a minimal one for this test run.
	planID := uuid.New()
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM public.plans WHERE id = $1`, planID)
	})
	_, err = db.ExecContext(ctx, `
		INSERT INTO public.plans (id, name, code, is_active, created_at, updated_at)
		VALUES ($1, 'Starter', 'starter', TRUE, NOW(), NOW())
		ON CONFLICT DO NOTHING
	`, planID)
	if err != nil {
		t.Fatalf("seed starter plan: %v", err)
	}
	return planID
}

// seedPricePlan inserts a plan with the given stripe price IDs and returns its planID.
func seedPricePlan(t *testing.T, db *sql.DB, ctx context.Context, monthlyPriceID, yearlyPriceID string) uuid.UUID {
	t.Helper()

	planID := uuid.New()
	suffix := planID.String()[:8]
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM public.plans WHERE id = $1`, planID)
	})
	_, err := db.ExecContext(ctx, `
		INSERT INTO public.plans
		       (id, name, code, is_active, stripe_price_id_monthly, stripe_price_id_yearly, created_at, updated_at)
		VALUES ($1, $2, $3, TRUE, $4, $5, NOW(), NOW())
	`, planID, "TestPlan "+suffix, "testplan_"+suffix, monthlyPriceID, yearlyPriceID)
	if err != nil {
		t.Fatalf("seed price plan: %v", err)
	}
	return planID
}

// fetchOrgPlan fetches the active organization_plans row for an org.
// Returns nil if none found.
func fetchOrgPlan(t *testing.T, db *sql.DB, ctx context.Context, orgID uuid.UUID) map[string]any {
	t.Helper()

	row := db.QueryRowContext(ctx, `
		SELECT plan_id::text, status, stripe_subscription_id, stripe_price_id,
		       billing_status, current_period_end
		FROM public.organization_plans
		WHERE organization_id = $1 AND status = 'active'
		LIMIT 1
	`, orgID)

	var planIDStr, status string
	var subID, priceID, billingStatus sql.NullString
	var periodEnd sql.NullTime
	if err := row.Scan(&planIDStr, &status, &subID, &priceID, &billingStatus, &periodEnd); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		t.Fatalf("fetchOrgPlan scan: %v", err)
	}
	return map[string]any{
		"plan_id":                planIDStr,
		"status":                 status,
		"stripe_subscription_id": subID.String,
		"stripe_price_id":        priceID.String,
		"billing_status":         billingStatus.String,
		"current_period_end":     periodEnd.Time,
	}
}

// ── Tests ──────────────────────────────────────────────────────────────────────

// TestPlanAssigner_Integration_AssignWithOrgID verifies the UPSERT path when
// OrgID is known: inserts an active organization_plans row with correct fields.
func TestPlanAssigner_Integration_AssignWithOrgID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := openPlanAssignerTestDB(t)
	defer db.Close()

	ctx := context.Background()
	orgID := seedTestOrg(t, db, ctx)
	monthlyPrice := fmt.Sprintf("price_inttest_%d_monthly", time.Now().UnixNano())
	yearlyPrice := fmt.Sprintf("price_inttest_%d_yearly", time.Now().UnixNano())
	planID := seedPricePlan(t, db, ctx, monthlyPrice, yearlyPrice)

	assigner := billingwiring.NewPlanAssigner(db)
	periodEnd := time.Now().UTC().Add(30 * 24 * time.Hour).Truncate(time.Second)

	err := assigner.Assign(ctx, services.AssignInput{
		OrgID:                orgID,
		StripeSubscriptionID: "sub_inttest_001",
		StripePriceID:        monthlyPrice,
		StripeCustomerID:     "cus_inttest_001",
		BillingStatus:        entities.BillingActive,
		CurrentPeriodEnd:     periodEnd,
		PaymentStatus:        "ok",
	})
	if err != nil {
		t.Fatalf("Assign: unexpected error: %v", err)
	}

	row := fetchOrgPlan(t, db, ctx, orgID)
	if row == nil {
		t.Fatal("expected active organization_plans row, got nil")
	}
	if row["plan_id"] != planID.String() {
		t.Errorf("plan_id: want %s, got %s", planID, row["plan_id"])
	}
	if row["stripe_subscription_id"] != "sub_inttest_001" {
		t.Errorf("stripe_subscription_id: want sub_inttest_001, got %q", row["stripe_subscription_id"])
	}
	if row["billing_status"] != "active" {
		t.Errorf("billing_status: want active, got %q", row["billing_status"])
	}
}

// TestPlanAssigner_Integration_AssignUpsert verifies that a second Assign call
// deactivates the first row and inserts a fresh active row.
func TestPlanAssigner_Integration_AssignUpsert(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := openPlanAssignerTestDB(t)
	defer db.Close()

	ctx := context.Background()
	orgID := seedTestOrg(t, db, ctx)
	basePrice := fmt.Sprintf("price_upsert_%d", time.Now().UnixNano())
	_ = seedPricePlan(t, db, ctx, basePrice+"_m", basePrice+"_y")
	newPrice := fmt.Sprintf("price_upsert_new_%d", time.Now().UnixNano())
	newPlanID := seedPricePlan(t, db, ctx, newPrice+"_m", newPrice+"_y")

	assigner := billingwiring.NewPlanAssigner(db)

	// First assignment.
	if err := assigner.Assign(ctx, services.AssignInput{
		OrgID:                orgID,
		StripeSubscriptionID: "sub_upsert_v1",
		StripePriceID:        basePrice + "_m",
		StripeCustomerID:     "cus_upsert_001",
		BillingStatus:        entities.BillingActive,
		CurrentPeriodEnd:     time.Now().UTC().Add(30 * 24 * time.Hour),
		PaymentStatus:        "ok",
	}); err != nil {
		t.Fatalf("Assign (first): %v", err)
	}

	// Second assignment — different plan, simulating an upgrade.
	if err := assigner.Assign(ctx, services.AssignInput{
		OrgID:                orgID,
		StripeSubscriptionID: "sub_upsert_v2",
		StripePriceID:        newPrice + "_m",
		StripeCustomerID:     "cus_upsert_001",
		BillingStatus:        entities.BillingActive,
		CurrentPeriodEnd:     time.Now().UTC().Add(60 * 24 * time.Hour),
		PaymentStatus:        "ok",
	}); err != nil {
		t.Fatalf("Assign (second): %v", err)
	}

	// Only one active row should exist with the NEW plan.
	row := fetchOrgPlan(t, db, ctx, orgID)
	if row == nil {
		t.Fatal("expected active organization_plans row after upsert, got nil")
	}
	if row["plan_id"] != newPlanID.String() {
		t.Errorf("plan_id after upsert: want %s, got %s", newPlanID, row["plan_id"])
	}
	if row["stripe_subscription_id"] != "sub_upsert_v2" {
		t.Errorf("stripe_subscription_id: want sub_upsert_v2, got %q", row["stripe_subscription_id"])
	}

	// Exactly one active row for this org.
	var activeCount int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM public.organization_plans WHERE organization_id = $1 AND status = 'active'`,
		orgID,
	).Scan(&activeCount); err != nil {
		t.Fatalf("active count query: %v", err)
	}
	if activeCount != 1 {
		t.Errorf("expected exactly 1 active org plan after upsert, got %d", activeCount)
	}
}

// TestPlanAssigner_Integration_OrphanCustomer verifies that when OrgID is nil
// and no org has the given stripe_customer_id, ErrOrphanCustomer is returned.
func TestPlanAssigner_Integration_OrphanCustomer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := openPlanAssignerTestDB(t)
	defer db.Close()

	ctx := context.Background()
	assigner := billingwiring.NewPlanAssigner(db)

	err := assigner.Assign(ctx, services.AssignInput{
		OrgID:            uuid.Nil,
		StripeCustomerID: fmt.Sprintf("cus_orphan_%d", time.Now().UnixNano()),
		StripePriceID:    "price_pro_monthly",
	})
	if err == nil {
		t.Fatal("expected ErrOrphanCustomer, got nil")
	}
	if err != billingwiring.ErrOrphanCustomer {
		t.Errorf("expected ErrOrphanCustomer, got: %v", err)
	}
}

// TestPlanAssigner_Integration_DowngradeToStarter verifies that when
// StripePriceID is empty, the assigner falls back to the starter plan.
func TestPlanAssigner_Integration_DowngradeToStarter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db := openPlanAssignerTestDB(t)
	defer db.Close()

	ctx := context.Background()
	orgID := seedTestOrg(t, db, ctx)
	starterPlanID := seedStarterPlan(t, db, ctx)

	assigner := billingwiring.NewPlanAssigner(db)

	// Downgrade: empty StripePriceID signals subscription.deleted.
	err := assigner.Assign(ctx, services.AssignInput{
		OrgID:            orgID,
		StripeCustomerID: "cus_downgrade_001",
		StripePriceID:    "", // empty → downgrade to starter
		BillingStatus:    entities.BillingCanceled,
	})
	if err != nil {
		t.Fatalf("Assign (downgrade): unexpected error: %v", err)
	}

	row := fetchOrgPlan(t, db, ctx, orgID)
	if row == nil {
		t.Fatal("expected active organization_plans row after downgrade, got nil")
	}
	if row["plan_id"] != starterPlanID.String() {
		t.Errorf("plan_id after downgrade: want starter %s, got %s", starterPlanID, row["plan_id"])
	}
	if row["billing_status"] != "canceled" {
		t.Errorf("billing_status: want canceled, got %q", row["billing_status"])
	}
}
