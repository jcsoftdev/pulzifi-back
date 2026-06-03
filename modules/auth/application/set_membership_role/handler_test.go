package setmembershiprole_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	setmembershiprole "github.com/jcsoftdev/pulzifi-back/modules/auth/application/set_membership_role"
)

type fakeMembershipWriter struct {
	isOwner bool
	ownErr  error
	setErr  error
	gotRole string
}

func (f *fakeMembershipWriter) IsOrgOwner(_ context.Context, _, _ uuid.UUID) (bool, error) {
	return f.isOwner, f.ownErr
}
func (f *fakeMembershipWriter) SetRole(_ context.Context, _, _ uuid.UUID, role string) error {
	f.gotRole = role
	return f.setErr
}

func TestSetRole_Valid(t *testing.T) {
	w := &fakeMembershipWriter{}
	h := setmembershiprole.NewHandler(w)
	err := h.Handle(context.Background(), setmembershiprole.Request{UserID: uuid.New(), OrgID: uuid.New(), Role: "ADMIN"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.gotRole != "ADMIN" {
		t.Errorf("role: got %q", w.gotRole)
	}
}

func TestSetRole_InvalidRoleRejected(t *testing.T) {
	h := setmembershiprole.NewHandler(&fakeMembershipWriter{})
	err := h.Handle(context.Background(), setmembershiprole.Request{UserID: uuid.New(), OrgID: uuid.New(), Role: "ROOT"})
	if err != setmembershiprole.ErrInvalidRole {
		t.Fatalf("want ErrInvalidRole, got %v", err)
	}
}

func TestSetRole_OwnerDemoteBlocked(t *testing.T) {
	h := setmembershiprole.NewHandler(&fakeMembershipWriter{isOwner: true})
	err := h.Handle(context.Background(), setmembershiprole.Request{UserID: uuid.New(), OrgID: uuid.New(), Role: "MEMBER"})
	if err != setmembershiprole.ErrCannotModifyOwner {
		t.Fatalf("want ErrCannotModifyOwner, got %v", err)
	}
}
