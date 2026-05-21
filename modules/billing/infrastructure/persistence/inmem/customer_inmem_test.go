package inmem

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/billing/domain/entities"
)

func TestCustomerRepo_SaveAndFindByOrgID(t *testing.T) {
	repo := NewCustomerRepo()
	orgID := uuid.New()
	customer := &entities.Customer{
		OrgID:            orgID,
		StripeCustomerID: "cus_abc123",
	}

	if err := repo.Save(context.Background(), customer); err != nil {
		t.Fatalf("Save: unexpected error: %v", err)
	}

	got, err := repo.FindByOrgID(context.Background(), orgID)
	if err != nil {
		t.Fatalf("FindByOrgID: unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("FindByOrgID: expected non-nil customer, got nil")
	}
	if got.StripeCustomerID != customer.StripeCustomerID {
		t.Errorf("StripeCustomerID: want %q, got %q", customer.StripeCustomerID, got.StripeCustomerID)
	}
}

func TestCustomerRepo_FindByOrgID_NotFound(t *testing.T) {
	repo := NewCustomerRepo()

	got, err := repo.FindByOrgID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for unknown orgID, got %+v", got)
	}
}

func TestCustomerRepo_FindByStripeCustomerID(t *testing.T) {
	repo := NewCustomerRepo()
	orgID := uuid.New()
	stripeID := "cus_xyz789"
	customer := &entities.Customer{OrgID: orgID, StripeCustomerID: stripeID}

	_ = repo.Save(context.Background(), customer)

	got, err := repo.FindByStripeCustomerID(context.Background(), stripeID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil customer")
	}
	if got.OrgID != orgID {
		t.Errorf("OrgID: want %v, got %v", orgID, got.OrgID)
	}
}

func TestCustomerRepo_ReturnsCopy(t *testing.T) {
	repo := NewCustomerRepo()
	orgID := uuid.New()
	original := &entities.Customer{OrgID: orgID, StripeCustomerID: "cus_original"}
	_ = repo.Save(context.Background(), original)

	got, _ := repo.FindByOrgID(context.Background(), orgID)
	// Mutate the returned copy — must not affect stored value.
	got.StripeCustomerID = "cus_mutated"

	got2, _ := repo.FindByOrgID(context.Background(), orgID)
	if got2.StripeCustomerID != "cus_original" {
		t.Errorf("stored value was mutated via returned pointer; got %q", got2.StripeCustomerID)
	}
}
