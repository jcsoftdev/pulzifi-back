package getsubscription

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/infrastructure/persistence/inmem"
)

func TestGetSubscriptionHandler_Handle(t *testing.T) {
	orgID := uuid.New()
	planID := uuid.New()
	periodEnd := time.Now().Add(30 * 24 * time.Hour).UTC().Truncate(time.Second)

	t.Run("returns subscription when org has an active Stripe subscription", func(t *testing.T) {
		repo := inmem.NewSubscriptionRepo()
		_ = repo.Save(context.Background(), &entities.Subscription{
			OrgID:                orgID,
			StripeSubscriptionID: "sub_test_123",
			StripeCustomerID:     "cus_test_abc",
			PlanID:               planID,
			BillingStatus:        entities.BillingActive,
			CurrentPeriodEnd:     &periodEnd,
			PaymentStatus:        "ok",
			UpdatedAt:            time.Now(),
		})

		h := NewHandler(repo)
		resp, err := h.Handle(context.Background(), orgID.String())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.OrgID != orgID.String() {
			t.Errorf("expected OrgID %q, got %q", orgID.String(), resp.OrgID)
		}
		if resp.BillingStatus != string(entities.BillingActive) {
			t.Errorf("expected billing status %q, got %q", entities.BillingActive, resp.BillingStatus)
		}
		if resp.StripeSubscriptionID != "sub_test_123" {
			t.Errorf("expected sub ID %q, got %q", "sub_test_123", resp.StripeSubscriptionID)
		}
		if resp.CurrentPeriodEnd == nil {
			t.Error("expected non-nil current_period_end")
		}
		if resp.PaymentStatus != "ok" {
			t.Errorf("expected payment status %q, got %q", "ok", resp.PaymentStatus)
		}
	})

	t.Run("returns ErrSubscriptionNotFound when org has no subscription", func(t *testing.T) {
		repo := inmem.NewSubscriptionRepo() // empty
		h := NewHandler(repo)

		_, err := h.Handle(context.Background(), orgID.String())

		if !errors.Is(err, ErrSubscriptionNotFound) {
			t.Errorf("expected ErrSubscriptionNotFound, got %v", err)
		}
	})

	t.Run("returns ErrSubscriptionNotFound for invalid org UUID", func(t *testing.T) {
		repo := inmem.NewSubscriptionRepo()
		h := NewHandler(repo)

		_, err := h.Handle(context.Background(), "not-a-uuid")

		if !errors.Is(err, ErrSubscriptionNotFound) {
			t.Errorf("expected ErrSubscriptionNotFound for bad UUID, got %v", err)
		}
	})

	// MUST: repo DB error is propagated to the caller with context.
	t.Run("propagates repository error when DB is unavailable", func(t *testing.T) {
		dbErr := errors.New("db connection lost")
		repo := &errorSubscriptionRepo{findErr: dbErr}
		h := NewHandler(repo)

		_, err := h.Handle(context.Background(), orgID.String())

		if err == nil {
			t.Fatal("expected error from failing repo, got nil")
		}
		// The handler must NOT swallow the raw repo error; it should be in the chain.
		if !errors.Is(err, dbErr) {
			t.Errorf("expected error chain to contain %v, got %v", dbErr, err)
		}
	})
}

// ── errorSubscriptionRepo is a hand-rolled stub that always returns an error ──
// It satisfies repositories.SubscriptionRepository.

type errorSubscriptionRepo struct {
	findErr error
}

func (r *errorSubscriptionRepo) FindByOrgID(ctx context.Context, orgID uuid.UUID) (*entities.Subscription, error) {
	return nil, r.findErr
}

func (r *errorSubscriptionRepo) FindByStripeSubscriptionID(ctx context.Context, subID string) (*entities.Subscription, error) {
	return nil, r.findErr
}

func (r *errorSubscriptionRepo) Save(ctx context.Context, sub *entities.Subscription) error {
	return r.findErr
}

func (r *errorSubscriptionRepo) Update(ctx context.Context, sub *entities.Subscription) error {
	return r.findErr
}
