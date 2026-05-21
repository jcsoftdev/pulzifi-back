package inmem

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
)

func TestSubscriptionRepo_SaveAndFindByOrgID(t *testing.T) {
	repo := NewSubscriptionRepo()
	orgID := uuid.New()
	end := time.Now().Add(30 * 24 * time.Hour)
	sub := &entities.Subscription{
		OrgID:                orgID,
		StripeSubscriptionID: "sub_123",
		StripeCustomerID:     "cus_123",
		BillingStatus:        "active",
		CurrentPeriodEnd:     &end,
	}

	if err := repo.Save(context.Background(), sub); err != nil {
		t.Fatalf("Save: unexpected error: %v", err)
	}

	got, err := repo.FindByOrgID(context.Background(), orgID)
	if err != nil {
		t.Fatalf("FindByOrgID: unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil subscription")
	}
	if got.StripeSubscriptionID != sub.StripeSubscriptionID {
		t.Errorf("StripeSubscriptionID: want %q, got %q", sub.StripeSubscriptionID, got.StripeSubscriptionID)
	}
}

func TestSubscriptionRepo_FindByOrgID_NotFound(t *testing.T) {
	repo := NewSubscriptionRepo()

	got, err := repo.FindByOrgID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for unknown orgID, got %+v", got)
	}
}

func TestSubscriptionRepo_Update_ChangesStoredRecord(t *testing.T) {
	repo := NewSubscriptionRepo()
	orgID := uuid.New()
	sub := &entities.Subscription{
		OrgID:                orgID,
		StripeSubscriptionID: "sub_abc",
		BillingStatus:        "active",
	}
	_ = repo.Save(context.Background(), sub)

	sub.BillingStatus = "canceled"
	sub.StripeSubscriptionID = "sub_abc_new"
	if err := repo.Update(context.Background(), sub); err != nil {
		t.Fatalf("Update: unexpected error: %v", err)
	}

	got, _ := repo.FindByOrgID(context.Background(), orgID)
	if got.BillingStatus != "canceled" {
		t.Errorf("BillingStatus after update: want 'canceled', got %q", got.BillingStatus)
	}
	if got.StripeSubscriptionID != "sub_abc_new" {
		t.Errorf("StripeSubscriptionID after update: want 'sub_abc_new', got %q", got.StripeSubscriptionID)
	}
}

func TestSubscriptionRepo_FindByStripeSubscriptionID(t *testing.T) {
	repo := NewSubscriptionRepo()
	orgID := uuid.New()
	sub := &entities.Subscription{
		OrgID:                orgID,
		StripeSubscriptionID: "sub_find_me",
	}
	_ = repo.Save(context.Background(), sub)

	got, err := repo.FindByStripeSubscriptionID(context.Background(), "sub_find_me")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil subscription")
	}
	if got.OrgID != orgID {
		t.Errorf("OrgID: want %v, got %v", orgID, got.OrgID)
	}
}
