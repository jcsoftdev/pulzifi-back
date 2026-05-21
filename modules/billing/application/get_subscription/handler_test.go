package getsubscription

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
)

// fakeSubscriptionRepo is a local in-memory SubscriptionRepository for this test file.
// It avoids importing modules/billing/infrastructure/persistence/inmem from the
// application layer (which violates the application→infrastructure boundary).
type fakeSubscriptionRepo struct {
	mu    sync.RWMutex
	byID  map[uuid.UUID]*entities.Subscription
	bySub map[string]*entities.Subscription
}

func newFakeSubscriptionRepo() *fakeSubscriptionRepo {
	return &fakeSubscriptionRepo{
		byID:  make(map[uuid.UUID]*entities.Subscription),
		bySub: make(map[string]*entities.Subscription),
	}
}

func (r *fakeSubscriptionRepo) FindByOrgID(_ context.Context, orgID uuid.UUID) (*entities.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if sub, ok := r.byID[orgID]; ok {
		cp := *sub
		return &cp, nil
	}
	return nil, nil
}

func (r *fakeSubscriptionRepo) FindByStripeSubscriptionID(_ context.Context, subID string) (*entities.Subscription, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if sub, ok := r.bySub[subID]; ok {
		cp := *sub
		return &cp, nil
	}
	return nil, nil
}

func (r *fakeSubscriptionRepo) Save(_ context.Context, sub *entities.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *sub
	r.byID[sub.OrgID] = &cp
	if sub.StripeSubscriptionID != "" {
		r.bySub[sub.StripeSubscriptionID] = &cp
	}
	return nil
}

func (r *fakeSubscriptionRepo) Update(_ context.Context, sub *entities.Subscription) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if old, ok := r.byID[sub.OrgID]; ok && old.StripeSubscriptionID != sub.StripeSubscriptionID {
		delete(r.bySub, old.StripeSubscriptionID)
	}
	cp := *sub
	r.byID[sub.OrgID] = &cp
	if sub.StripeSubscriptionID != "" {
		r.bySub[sub.StripeSubscriptionID] = &cp
	}
	return nil
}

func TestGetSubscriptionHandler_Handle(t *testing.T) {
	orgID := uuid.New()
	planID := uuid.New()
	periodEnd := time.Now().Add(30 * 24 * time.Hour).UTC().Truncate(time.Second)

	t.Run("returns subscription when org has an active Stripe subscription", func(t *testing.T) {
		repo := newFakeSubscriptionRepo()
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
		repo := newFakeSubscriptionRepo() // empty
		h := NewHandler(repo)

		_, err := h.Handle(context.Background(), orgID.String())

		if !errors.Is(err, ErrSubscriptionNotFound) {
			t.Errorf("expected ErrSubscriptionNotFound, got %v", err)
		}
	})

	t.Run("returns ErrSubscriptionNotFound for invalid org UUID", func(t *testing.T) {
		repo := newFakeSubscriptionRepo()
		h := NewHandler(repo)

		_, err := h.Handle(context.Background(), "not-a-uuid")

		if !errors.Is(err, ErrSubscriptionNotFound) {
			t.Errorf("expected ErrSubscriptionNotFound for bad UUID, got %v", err)
		}
	})
}
