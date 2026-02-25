package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/entities"
)

type TeamMemberRepository interface {
	ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*entities.TeamMember, error)
	GetByID(ctx context.Context, memberID uuid.UUID) (*entities.TeamMember, error)
	GetByUserAndOrg(ctx context.Context, orgID, userID uuid.UUID) (*entities.TeamMember, error)
	FindUserByEmail(ctx context.Context, email string) (*entities.TeamMember, error)
	CreateUser(ctx context.Context, email, firstName, lastName, hashedPassword string) (uuid.UUID, error)
	AddMember(ctx context.Context, orgID, userID uuid.UUID, role string, invitedBy *uuid.UUID, invitationStatus string) (*entities.TeamMember, error)
	UpdateRole(ctx context.Context, memberID uuid.UUID, role string) error
	UpdateInvitationStatus(ctx context.Context, memberID uuid.UUID, status string) error
	Remove(ctx context.Context, memberID uuid.UUID) error
	GetOrganizationIDBySubdomain(ctx context.Context, subdomain string) (uuid.UUID, error)
}
