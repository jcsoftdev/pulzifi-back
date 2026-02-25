package listmembers

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/repositories"
)

type ListMembersHandler struct {
	repo repositories.TeamMemberRepository
}

func NewListMembersHandler(repo repositories.TeamMemberRepository) *ListMembersHandler {
	return &ListMembersHandler{repo: repo}
}

func (h *ListMembersHandler) Handle(ctx context.Context, subdomain string) (*ListMembersResponse, error) {
	orgID, err := h.repo.GetOrganizationIDBySubdomain(ctx, subdomain)
	if err != nil {
		return nil, err
	}

	members, err := h.repo.ListByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*TeamMemberResponse, 0, len(members))
	for _, m := range members {
		dtos = append(dtos, &TeamMemberResponse{
			ID:               m.ID,
			UserID:           m.UserID,
			Role:             m.Role,
			FirstName:        m.FirstName,
			LastName:         m.LastName,
			Email:            m.Email,
			AvatarURL:        m.AvatarURL,
			InvitedBy:        m.InvitedBy,
			JoinedAt:         m.JoinedAt,
			InvitationStatus: m.InvitationStatus,
		})
	}

	return &ListMembersResponse{Members: dtos}, nil
}
