package billingwiring_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"

	billingwiring "github.com/jcsoftdev/pulzifi-back/cmd/wiring/billing"
)

// ── fakeTxDB ─────────────────────────────────────────────────────────────────
// fakeTxDB implements billingwiring.TxDB, the minimal interface the PlanAssigner
// needs from *sql.DB.  This lets us unit-test the adapter without Postgres.

type fakeTxDB struct {
	tx     *fakeTx
	txErr  error
}

func (f *fakeTxDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (billingwiring.Tx, error) {
	if f.txErr != nil {
		return nil, f.txErr
	}
	return f.tx, nil
}

// ── fakeTx ───────────────────────────────────────────────────────────────────

type fakeTx struct {
	// Expectations for QueryRow calls (in order)
	rows      []*fakeRow
	rowIdx    int
	execErr   error
	commitErr error
	rolledBack bool
	committed  bool
}

func (t *fakeTx) QueryRowContext(_ context.Context, _ string, _ ...any) billingwiring.RowScanner {
	if t.rowIdx >= len(t.rows) {
		return &fakeRow{err: sql.ErrNoRows}
	}
	row := t.rows[t.rowIdx]
	t.rowIdx++
	return row
}

func (t *fakeTx) ExecContext(_ context.Context, _ string, _ ...any) (sql.Result, error) {
	return fakeResult{}, t.execErr
}

func (t *fakeTx) Commit() error {
	t.committed = true
	return t.commitErr
}

func (t *fakeTx) Rollback() error {
	t.rolledBack = true
	return nil
}

// ── fakeRow ───────────────────────────────────────────────────────────────────

type fakeRow struct {
	val string
	err error
}

func (r *fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) > 0 {
		if s, ok := dest[0].(*string); ok {
			*s = r.val
		}
	}
	return nil
}

// ── fakeResult ───────────────────────────────────────────────────────────────

type fakeResult struct{}

func (fakeResult) LastInsertId() (int64, error) { return 0, nil }
func (fakeResult) RowsAffected() (int64, error)  { return 1, nil }

// ── helpers ───────────────────────────────────────────────────────────────────

func uuidRow(id uuid.UUID) *fakeRow { return &fakeRow{val: id.String()} }
func noRow() *fakeRow               { return &fakeRow{err: sql.ErrNoRows} }
func errRow(err error) *fakeRow     { return &fakeRow{err: err} }

// ── Tests ─────────────────────────────────────────────────────────────────────

// TestAssign_OrgIDProvided_HappyPath: org is known; price resolves to a plan; upsert succeeds.
func TestAssign_OrgIDProvided_HappyPath(t *testing.T) {
	planID := uuid.New()
	priceID := "price_pro_monthly"

	tx := &fakeTx{
		// First QueryRow: plan lookup by priceID → returns planID
		rows: []*fakeRow{uuidRow(planID)},
	}
	db := &fakeTxDB{tx: tx}
	a := billingwiring.NewPlanAssignerWithDB(db)

	err := a.Assign(context.Background(), services.AssignInput{
		OrgID:                uuid.New(),
		StripeSubscriptionID: "sub_abc",
		StripePriceID:        priceID,
		StripeCustomerID:     "cus_abc",
		BillingStatus:        entities.BillingActive,
		CurrentPeriodEnd:     time.Now().Add(30 * 24 * time.Hour),
		PaymentStatus:        "ok",
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if !tx.committed {
		t.Error("expected transaction to be committed")
	}
}

// TestAssign_OrgIDNil_CustomerLookup: OrgID is uuid.Nil, assigner resolves via StripeCustomerID.
func TestAssign_OrgIDNil_CustomerLookup(t *testing.T) {
	orgID := uuid.New()
	planID := uuid.New()
	customerID := "cus_xyz"

	tx := &fakeTx{
		rows: []*fakeRow{
			// First QueryRow: org lookup by stripe_customer_id → returns orgID
			uuidRow(orgID),
			// Second QueryRow: plan lookup by priceID → returns planID
			uuidRow(planID),
		},
	}
	db := &fakeTxDB{tx: tx}
	a := billingwiring.NewPlanAssignerWithDB(db)

	err := a.Assign(context.Background(), services.AssignInput{
		OrgID:                uuid.Nil,
		StripeSubscriptionID: "sub_def",
		StripePriceID:        "price_pro_yearly",
		StripeCustomerID:     customerID,
		BillingStatus:        entities.BillingActive,
		CurrentPeriodEnd:     time.Now().Add(365 * 24 * time.Hour),
		PaymentStatus:        "ok",
	})
	if err != nil {
		t.Fatalf("expected nil error on customer lookup path, got: %v", err)
	}
	if !tx.committed {
		t.Error("expected transaction to be committed")
	}
}

// TestAssign_OrphanCustomer: OrgID is nil and customer has no matching org.
func TestAssign_OrphanCustomer(t *testing.T) {
	tx := &fakeTx{
		// First QueryRow: org lookup by stripe_customer_id → no rows
		rows: []*fakeRow{noRow()},
	}
	db := &fakeTxDB{tx: tx}
	a := billingwiring.NewPlanAssignerWithDB(db)

	err := a.Assign(context.Background(), services.AssignInput{
		OrgID:            uuid.Nil,
		StripeCustomerID: "cus_unknown",
		StripePriceID:    "price_pro_monthly",
	})
	if err == nil {
		t.Fatal("expected ErrOrphanCustomer, got nil")
	}
	if !errors.Is(err, billingwiring.ErrOrphanCustomer) {
		t.Fatalf("expected ErrOrphanCustomer, got: %v", err)
	}
	if !tx.rolledBack {
		t.Error("expected rollback on orphan customer error")
	}
}

// TestAssign_PlanNotFound: price ID maps to no plan in DB.
func TestAssign_PlanNotFound(t *testing.T) {
	orgID := uuid.New()

	tx := &fakeTx{
		// First QueryRow: plan lookup → no rows (price not in DB)
		rows: []*fakeRow{noRow()},
	}
	db := &fakeTxDB{tx: tx}
	a := billingwiring.NewPlanAssignerWithDB(db)

	err := a.Assign(context.Background(), services.AssignInput{
		OrgID:         orgID,
		StripePriceID: "price_unknown",
	})
	if err == nil {
		t.Fatal("expected error for unknown price ID, got nil")
	}
	if !errors.Is(err, billingwiring.ErrPlanNotFound) {
		t.Fatalf("expected ErrPlanNotFound, got: %v", err)
	}
}

// TestAssign_EmptyPriceID_Downgrade: When StripePriceID is empty, the assigner
// falls back to the starter plan (subscription.deleted / cancellation path).
func TestAssign_EmptyPriceID_Downgrade(t *testing.T) {
	orgID := uuid.New()
	starterPlanID := uuid.New()

	tx := &fakeTx{
		// First QueryRow: plan lookup uses "starter" code path → returns starterPlanID
		rows: []*fakeRow{uuidRow(starterPlanID)},
	}
	db := &fakeTxDB{tx: tx}
	a := billingwiring.NewPlanAssignerWithDB(db)

	err := a.Assign(context.Background(), services.AssignInput{
		OrgID:            orgID,
		StripePriceID:    "", // empty — downgrade signal
		StripeCustomerID: "cus_def",
		BillingStatus:    entities.BillingCanceled,
	})
	if err != nil {
		t.Fatalf("expected nil error on downgrade path, got: %v", err)
	}
	if !tx.committed {
		t.Error("expected transaction to be committed on downgrade")
	}
}
