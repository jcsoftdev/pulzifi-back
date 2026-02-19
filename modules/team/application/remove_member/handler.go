package removemember

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/repositories"
)

type RemoveMemberHandler struct {
	repo repositories.TeamMemberRepository
}

func NewRemoveMemberHandler(repo repositories.TeamMemberRepository) *RemoveMemberHandler {
	return &RemoveMemberHandler{repo: repo}
}

func (h *RemoveMemberHandler) Handle(ctx context.Context, memberID, requesterID uuid.UUID) error {
	member, err := h.repo.GetByID(ctx, memberID)
	if err != nil || member == nil {
		return ErrMemberNotFound
	}

	if member.Role == "OWNER" {
		return ErrCannotRemoveOwner
	}

	if member.UserID == requesterID {
		return ErrCannotRemoveSelf
	}

	return h.repo.Remove(ctx, memberID)
}
