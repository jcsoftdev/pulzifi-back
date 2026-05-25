// Package billingwiring provides the cross-module adapter that implements
// modules/billing/domain/services.PlanAssigner. It runs raw SQL against the
// public schema so the billing module never imports usage or organization
// infrastructure packages directly.
package billingwiring

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/services"
)

// ErrOrphanCustomer is the wiring-side alias of services.ErrOrphanCustomer.
// Both values compare equal under errors.Is so the HTTP layer can ack 200
// without importing the wiring package directly.
var ErrOrphanCustomer = services.ErrOrphanCustomer

// ErrPlanNotFound re-exports the domain sentinel so callers that import this
// package for wiring can still use errors.Is checks.  The canonical definition
// lives in modules/billing/domain/services to keep the application layer clean.
var ErrPlanNotFound = services.ErrPlanNotFound

// RowScanner is the minimal interface over *sql.Row used by the adapter.
// It exists so tests can inject fakes without a real database.
type RowScanner interface {
	Scan(dest ...any) error
}

// Tx is the minimal interface over *sql.Tx needed by PlanAssigner.
type Tx interface {
	QueryRowContext(ctx context.Context, query string, args ...any) RowScanner
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Commit() error
	Rollback() error
}

// TxDB is the minimal interface over *sql.DB that PlanAssigner needs.
// The concrete implementation (sqlTxDB) wraps *sql.DB; tests inject fakes.
type TxDB interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
}

// ── Compile-time interface check ──────────────────────────────────────────────

var _ services.PlanAssigner = (*PlanAssigner)(nil)

// ── PlanAssigner ──────────────────────────────────────────────────────────────

// PlanAssigner implements services.PlanAssigner.
// It uses raw SQL on the public schema — no SET search_path needed.
type PlanAssigner struct {
	db TxDB
}

// NewPlanAssigner constructs a PlanAssigner backed by a real *sql.DB.
// This is what cmd/server/modules.go calls.
func NewPlanAssigner(db *sql.DB) *PlanAssigner {
	return &PlanAssigner{db: &sqlTxDB{db: db}}
}

// NewPlanAssignerWithDB constructs a PlanAssigner with an injected TxDB.
// Used by tests to inject fakes.
func NewPlanAssignerWithDB(db TxDB) *PlanAssigner {
	return &PlanAssigner{db: db}
}

// Assign resolves the plan from the Stripe price ID and upserts the
// organization_plans row for the organisation.
//
// If in.OrgID == uuid.Nil the adapter first resolves the org by looking up
// in.StripeCustomerID in public.organizations.stripe_customer_id.
// If no org is found, ErrOrphanCustomer is returned.
//
// The upsert semantics align with the usage module's plan-assignment path:
// deactivate any existing active row, then insert a fresh one with the new
// Stripe data.  This keeps the table-level approach consistent until the two
// code paths are consolidated.
func (a *PlanAssigner) Assign(ctx context.Context, in services.AssignInput) error {
	tx, err := a.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("billing plan assigner: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1. Resolve org ID if not provided
	orgID := in.OrgID
	if orgID == uuid.Nil {
		orgID, err = a.resolveOrgByCustomer(ctx, tx, in.StripeCustomerID)
		if err != nil {
			return err
		}
	}

	// 2. Resolve plan ID from stripe price ID.
	//    When StripePriceID is empty (subscription.deleted → downgrade), look up
	//    the starter plan by code so we can safely reset the org.
	planID, err := a.resolvePlanID(ctx, tx, in.StripePriceID)
	if err != nil {
		return err
	}

	// 3. Deactivate existing active plan row (mirrors usage module's approach).
	_, err = tx.ExecContext(ctx, `
		UPDATE public.organization_plans
		SET    status     = 'inactive',
		       ended_at   = NOW(),
		       updated_at = NOW()
		WHERE  organization_id = $1
		  AND  status          = 'active'
		  AND  deleted_at      IS NULL
	`, orgID)
	if err != nil {
		return fmt.Errorf("billing plan assigner: deactivate old plan: %w", err)
	}

	// 4. Insert new active plan row with Stripe billing fields.
	billingStatus := string(in.BillingStatus)
	if billingStatus == "" {
		billingStatus = string(entities.BillingActive)
	}
	paymentStatus := in.PaymentStatus
	if paymentStatus == "" {
		paymentStatus = "ok"
	}

	var periodEnd *time.Time
	if !in.CurrentPeriodEnd.IsZero() {
		t := in.CurrentPeriodEnd
		periodEnd = &t
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO public.organization_plans
		       (organization_id, plan_id, status, started_at,
		        stripe_subscription_id, stripe_price_id, billing_status,
		        current_period_end, payment_status,
		        created_at, updated_at)
		VALUES ($1, $2, 'active', NOW(),
		        $3, $4, $5,
		        $6, $7,
		        NOW(), NOW())
	`, orgID, planID,
		in.StripeSubscriptionID, in.StripePriceID, billingStatus,
		periodEnd, paymentStatus,
	)
	if err != nil {
		return fmt.Errorf("billing plan assigner: insert plan: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("billing plan assigner: commit: %w", err)
	}
	return nil
}

// resolveOrgByCustomer looks up the organization ID from the Stripe customer ID.
func (a *PlanAssigner) resolveOrgByCustomer(ctx context.Context, tx Tx, customerID string) (uuid.UUID, error) {
	var orgIDStr string
	row := tx.QueryRowContext(ctx, `
		SELECT id::text
		  FROM public.organizations
		 WHERE stripe_customer_id = $1
		   AND deleted_at IS NULL
		 LIMIT 1
	`, customerID)
	if err := row.Scan(&orgIDStr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, ErrOrphanCustomer
		}
		return uuid.Nil, fmt.Errorf("billing plan assigner: resolve org by customer: %w", err)
	}
	id, err := uuid.Parse(orgIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("billing plan assigner: parse org UUID: %w", err)
	}
	return id, nil
}

// resolvePlanID maps a Stripe price ID to an internal plan UUID.
// When priceID is empty (downgrade / subscription.deleted), it falls back to
// the plan with code = 'starter'.
func (a *PlanAssigner) resolvePlanID(ctx context.Context, tx Tx, priceID string) (uuid.UUID, error) {
	var planIDStr string
	var row RowScanner

	if priceID == "" {
		// Downgrade path — use starter plan
		row = tx.QueryRowContext(ctx, `
			SELECT id::text
			  FROM public.plans
			 WHERE code = 'starter'
			   AND is_active = TRUE
			 LIMIT 1
		`)
	} else {
		row = tx.QueryRowContext(ctx, `
			SELECT id::text
			  FROM public.plans
			 WHERE (stripe_price_id_monthly = $1 OR stripe_price_id_yearly = $1)
			   AND is_active = TRUE
			 LIMIT 1
		`, priceID)
	}

	if err := row.Scan(&planIDStr); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, ErrPlanNotFound
		}
		return uuid.Nil, fmt.Errorf("billing plan assigner: resolve plan: %w", err)
	}
	id, err := uuid.Parse(planIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("billing plan assigner: parse plan UUID: %w", err)
	}
	return id, nil
}

// ── sqlTxDB ───────────────────────────────────────────────────────────────────

// sqlTxDB wraps *sql.DB to implement TxDB, adapting *sql.Tx to our Tx interface.
type sqlTxDB struct{ db *sql.DB }

func (s *sqlTxDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	tx, err := s.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &sqlTx{tx: tx}, nil
}

// sqlTx wraps *sql.Tx and implements our Tx interface by wrapping QueryRowContext
// to return our RowScanner interface instead of *sql.Row.
type sqlTx struct{ tx *sql.Tx }

func (t *sqlTx) QueryRowContext(ctx context.Context, query string, args ...any) RowScanner {
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *sqlTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *sqlTx) Commit() error   { return t.tx.Commit() }
func (t *sqlTx) Rollback() error { return t.tx.Rollback() }
