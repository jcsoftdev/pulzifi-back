package listinvitations

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/admin/domain/repositories"
)

type Handler struct {
	repo repositories.InvitationRepository
}

func New(repo repositories.InvitationRepository) *Handler { return &Handler{repo: repo} }

func (h *Handler) Handle(ctx context.Context, status string, limit, offset int) (*Response, error) {
	invs, total, err := h.repo.List(ctx, repositories.ListInvitationsFilter{Status: status, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	out := make([]InvitationDTO, 0, len(invs))
	for _, inv := range invs {
		out = append(out, toDTO(inv))
	}
	return &Response{Invitations: out, Total: total}, nil
}

func toDTO(inv *entities.Invitation) InvitationDTO {
	return InvitationDTO{
		ID:          inv.ID.String(),
		Email:       inv.Email,
		Status:      string(inv.Status),
		InvitedBy:   inv.InvitedBy.String(),
		ExpiresAt:   inv.ExpiresAt,
		EmailSentAt: inv.EmailSentAt,
		EmailError:  inv.EmailError,
		ResentCount: inv.ResentCount,
		CreatedAt:   inv.CreatedAt,
	}
}
