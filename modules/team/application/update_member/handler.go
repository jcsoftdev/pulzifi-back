package updatemember

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/repositories"
)

type UpdateMemberHandler struct {
	repo repositories.TeamMemberRepository
}

func NewUpdateMemberHandler(repo repositories.TeamMemberRepository) *UpdateMemberHandler {
	return &UpdateMemberHandler{repo: repo}
}

func (h *UpdateMemberHandler) Handle(ctx context.Context, memberID uuid.UUID, req *UpdateMemberRequest) error {
	member, err := h.repo.GetByID(ctx, memberID)
	if err != nil || member == nil {
		return ErrMemberNotFound
	}

	if member.Role == "OWNER" {
		return ErrCannotUpdateOwnerRole
	}

	role := strings.ToUpper(req.Role)
	if role == "" {
		role = "MEMBER"
	}

	return h.repo.UpdateRole(ctx, memberID, role)
}
