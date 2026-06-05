//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping integration test")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return db
}

func TestPlanPostgres_UpsertStripePricing(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	ctx := context.Background()
	repo := NewPlanPostgresRepository(db)

	_, err := db.ExecContext(ctx, `
		INSERT INTO public.plans (code, name, checks_allowed_monthly, is_active)
		VALUES ('test_pro', 'Test Pro', 100, TRUE)
		ON CONFLICT (code) DO UPDATE SET is_active = TRUE`)
	if err != nil {
		t.Fatalf("seed plan: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, `DELETE FROM public.plans WHERE code = 'test_pro'`) })

	if err := repo.UpsertStripePricing(ctx, "test_pro", "monthly", "price_m_1", 2000, "usd"); err != nil {
		t.Fatalf("upsert monthly: %v", err)
	}
	if err := repo.UpsertStripePricing(ctx, "test_pro", "yearly", "price_y_1", 19200, "usd"); err != nil {
		t.Fatalf("upsert yearly: %v", err)
	}

	var (
		pm, py sql.NullString
		am, ay sql.NullInt64
		cur    string
	)
	err = db.QueryRowContext(ctx, `
		SELECT stripe_price_id_monthly, stripe_price_id_yearly,
		       amount_cents_monthly, amount_cents_yearly, currency
		FROM public.plans WHERE code = 'test_pro'`).Scan(&pm, &py, &am, &ay, &cur)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if pm.String != "price_m_1" || am.Int64 != 2000 {
		t.Errorf("monthly: got (%q, %d), want (price_m_1, 2000)", pm.String, am.Int64)
	}
	if py.String != "price_y_1" || ay.Int64 != 19200 {
		t.Errorf("yearly: got (%q, %d), want (price_y_1, 19200)", py.String, ay.Int64)
	}
	if cur != "usd" {
		t.Errorf("currency: got %q, want usd", cur)
	}

	err = repo.UpsertStripePricing(ctx, "does_not_exist", "monthly", "price_x", 100, "usd")
	if err == nil {
		t.Error("expected error for unknown plan code, got nil")
	}
}
