package invitemember

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/repositories"
)

type InviteMemberHandler struct {
	repo repositories.TeamMemberRepository
}

func NewInviteMemberHandler(repo repositories.TeamMemberRepository) *InviteMemberHandler {
	return &InviteMemberHandler{repo: repo}
}

func (h *InviteMemberHandler) Handle(ctx context.Context, subdomain string, inviterID uuid.UUID, req *InviteMemberRequest) (*InviteMemberResponse, error) {
	role := strings.ToUpper(req.Role)
	if role == "" {
		role = "MEMBER"
	}

	orgID, err := h.repo.GetOrganizationIDBySubdomain(ctx, subdomain)
	if err != nil {
		return nil, ErrOrganizationNotFound
	}

	// Find user by email
	user, err := h.repo.FindUserByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}

	// Check if already member
	existing, err := h.repo.GetByUserAndOrg(ctx, orgID, user.UserID)
	if err == nil && existing != nil {
		return nil, ErrAlreadyMember
	}

	inviterIDPtr := &inviterID
	member, err := h.repo.AddMember(ctx, orgID, user.UserID, role, inviterIDPtr)
	if err != nil {
		return nil, err
	}

	return &InviteMemberResponse{
		ID:        member.ID,
		UserID:    member.UserID,
		Role:      member.Role,
		FirstName: member.FirstName,
		LastName:  member.LastName,
		Email:     member.Email,
		JoinedAt:  member.JoinedAt,
	}, nil
}
