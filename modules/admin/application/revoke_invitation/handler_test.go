package revokeinvitation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	revokeinvitation "github.com/jcsoftdev/pulzifi-back/modules/admin/application/revoke_invitation"
	domerrs "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/errors"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories/mocks"
)

func TestRevoke_Pending_NoError(t *testing.T) {
	id := uuid.New()
	revokedBy := uuid.New()

	repo := &mocks.InvitationRepositoryMock{
		RevokeFunc: func(ctx context.Context, gotID uuid.UUID, gotRevokedBy uuid.UUID) error {
			if gotID != id {
				t.Fatalf("ID mismatch")
			}
			if gotRevokedBy != revokedBy {
				t.Fatalf("revokedBy mismatch")
			}
			return nil
		},
	}

	h := revokeinvitation.New(repo)
	if err := h.Handle(context.Background(), id, revokedBy); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRevoke_AlreadyDecided_Propagates(t *testing.T) {
	repo := &mocks.InvitationRepositoryMock{
		RevokeFunc: func(ctx context.Context, id uuid.UUID, revokedBy uuid.UUID) error {
			return domerrs.ErrInvitationAlreadyDecided
		},
	}

	h := revokeinvitation.New(repo)
	err := h.Handle(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domerrs.ErrInvitationAlreadyDecided) {
		t.Fatalf("expected ErrInvitationAlreadyDecided, got %v", err)
	}
}
