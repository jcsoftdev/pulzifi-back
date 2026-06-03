package removemembership_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	removemembership "github.com/jcsoftdev/pulzifi-back/modules/auth/application/remove_membership"
)

type fakeRemover struct {
	isOwner   bool
	removed   bool
	ownErr    error
	removeErr error
}

func (f *fakeRemover) IsOrgOwner(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return f.isOwner, f.ownErr
}
func (f *fakeRemover) RemoveMembership(_ context.Context, _, _ uuid.UUID) error {
	f.removed = true
	return f.removeErr
}

func TestRemove_Valid(t *testing.T) {
	f := &fakeRemover{}
	h := removemembership.NewHandler(f)
	err := h.Handle(context.Background(), removemembership.Request{UserID: uuid.New(), OrgID: uuid.New()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !f.removed {
		t.Error("expected RemoveMembership to be called")
	}
}

func TestRemove_OwnerBlocked(t *testing.T) {
	h := removemembership.NewHandler(&fakeRemover{isOwner: true})
	err := h.Handle(context.Background(), removemembership.Request{UserID: uuid.New(), OrgID: uuid.New()})
	if err != removemembership.ErrCannotRemoveOwner {
		t.Fatalf("want ErrCannotRemoveOwner, got %v", err)
	}
}
