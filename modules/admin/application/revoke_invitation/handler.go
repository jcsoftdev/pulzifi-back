package revokeinvitation

import (
	"context"

	"github.com/google/uuid"

	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
)

type Handler struct {
	repo repositories.InvitationRepository
}

func New(repo repositories.InvitationRepository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Handle(ctx context.Context, id uuid.UUID, revokedBy uuid.UUID) error {
	return h.repo.Revoke(ctx, id, revokedBy)
}
