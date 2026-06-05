package getuserdetail_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	getuserdetail "github.com/jcsoftdev/pulzifi-back/modules/auth/application/get_user_detail"
)

type fakeReader struct {
	detail *getuserdetail.UserDetail
	err    error
}

func (f *fakeReader) GetUserDetail(_ context.Context, _ uuid.UUID) (*getuserdetail.UserDetail, error) {
	return f.detail, f.err
}

func TestGetUserDetail_MapsMemberships(t *testing.T) {
	id := uuid.New()
	detail := &getuserdetail.UserDetail{
		ID: id, Email: "a@b.com", FirstName: "A", LastName: "B", Status: "approved", EmailVerified: true,
		Memberships: []getuserdetail.Membership{
			{OrgID: uuid.New(), OrgName: "Acme", Subdomain: "acme", Role: "OWNER", InvitationStatus: "active", IsOwner: true},
		},
	}
	h := getuserdetail.NewHandler(&fakeReader{detail: detail})

	resp, err := h.Handle(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Email != "a@b.com" {
		t.Errorf("email: got %q", resp.Email)
	}
	if len(resp.Memberships) != 1 || resp.Memberships[0].IsOwner != true {
		t.Fatalf("membership mapping failed: %+v", resp.Memberships)
	}
}

func TestGetUserDetail_NotFound(t *testing.T) {
	h := getuserdetail.NewHandler(&fakeReader{detail: nil, err: getuserdetail.ErrUserNotFound})
	_, err := h.Handle(context.Background(), uuid.New())
	if err != getuserdetail.ErrUserNotFound {
		t.Fatalf("want ErrUserNotFound, got %v", err)
	}
}
