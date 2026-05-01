package listinvitations_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	listinvitations "github.com/jcsoftdev/pulzifi-back/modules/admin/application/list_invitations"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories/mocks"
)

func TestList_ReturnsInvitations(t *testing.T) {
	inviteID := uuid.New()
	inviterID := uuid.New()
	now := time.Now().UTC()
	expires := now.Add(72 * time.Hour)

	repo := &mocks.InvitationRepositoryMock{
		ListFunc: func(ctx context.Context, f repositories.ListInvitationsFilter) ([]*entities.Invitation, int, error) {
			if f.Status != "pending" {
				t.Fatalf("expected status 'pending', got %s", f.Status)
			}
			if f.Limit != 50 {
				t.Fatalf("expected limit 50, got %d", f.Limit)
			}
			if f.Offset != 0 {
				t.Fatalf("expected offset 0, got %d", f.Offset)
			}
			return []*entities.Invitation{
				{
					ID:          inviteID,
					Email:       "foo@example.com",
					Status:      entities.InvitationPending,
					InvitedBy:   inviterID,
					ExpiresAt:   expires,
					ResentCount: 0,
					CreatedAt:   now,
				},
			}, 1, nil
		},
	}

	h := listinvitations.New(repo)
	resp, err := h.Handle(context.Background(), "pending", 50, 0)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Total != 1 {
		t.Fatalf("expected total=1, got %d", resp.Total)
	}
	if len(resp.Invitations) != 1 {
		t.Fatalf("expected 1 invitation, got %d", len(resp.Invitations))
	}
	got := resp.Invitations[0]
	if got.ID != inviteID.String() {
		t.Fatalf("ID mismatch: %s", got.ID)
	}
	if got.Email != "foo@example.com" {
		t.Fatalf("Email mismatch: %s", got.Email)
	}
	if got.Status != "pending" {
		t.Fatalf("Status mismatch: %s", got.Status)
	}
	if got.InvitedBy != inviterID.String() {
		t.Fatalf("InvitedBy mismatch: %s", got.InvitedBy)
	}
}
