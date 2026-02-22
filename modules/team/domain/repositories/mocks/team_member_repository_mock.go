package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/team/domain/entities"
)

type MockTeamMemberRepository struct {
	ListByOrganizationResult []*entities.TeamMember
	ListByOrganizationErr    error
	GetByIDResult            *entities.TeamMember
	GetByIDErr               error
	GetByUserAndOrgResult    *entities.TeamMember
	GetByUserAndOrgErr       error
	FindUserByEmailResult    *entities.TeamMember
	FindUserByEmailErr       error
	CreateUserResult         uuid.UUID
	CreateUserErr            error
	AddMemberResult          *entities.TeamMember
	AddMemberErr             error
	UpdateRoleErr            error
	RemoveErr                error
	GetOrgIDBySubdomainResult uuid.UUID
	GetOrgIDBySubdomainErr    error

	FindUserByEmailFn       func(ctx context.Context, email string) (*entities.TeamMember, error)
	GetByUserAndOrgFn       func(ctx context.Context, orgID, userID uuid.UUID) (*entities.TeamMember, error)
	AddMemberFn             func(ctx context.Context, orgID, userID uuid.UUID, role string, invitedBy *uuid.UUID) (*entities.TeamMember, error)
	GetOrgIDBySubdomainFn   func(ctx context.Context, subdomain string) (uuid.UUID, error)
	CreateUserFn            func(ctx context.Context, email, firstName, lastName, hashedPassword string) (uuid.UUID, error)

	FindUserByEmailCalls int
	CreateUserCalls      int
	AddMemberCalls       int
}

func (m *MockTeamMemberRepository) ListByOrganization(_ context.Context, _ uuid.UUID) ([]*entities.TeamMember, error) {
	return m.ListByOrganizationResult, m.ListByOrganizationErr
}

func (m *MockTeamMemberRepository) GetByID(_ context.Context, _ uuid.UUID) (*entities.TeamMember, error) {
	return m.GetByIDResult, m.GetByIDErr
}

func (m *MockTeamMemberRepository) GetByUserAndOrg(ctx context.Context, orgID, userID uuid.UUID) (*entities.TeamMember, error) {
	if m.GetByUserAndOrgFn != nil {
		return m.GetByUserAndOrgFn(ctx, orgID, userID)
	}
	return m.GetByUserAndOrgResult, m.GetByUserAndOrgErr
}

func (m *MockTeamMemberRepository) FindUserByEmail(ctx context.Context, email string) (*entities.TeamMember, error) {
	m.FindUserByEmailCalls++
	if m.FindUserByEmailFn != nil {
		return m.FindUserByEmailFn(ctx, email)
	}
	return m.FindUserByEmailResult, m.FindUserByEmailErr
}

func (m *MockTeamMemberRepository) CreateUser(ctx context.Context, email, firstName, lastName, hashedPassword string) (uuid.UUID, error) {
	m.CreateUserCalls++
	if m.CreateUserFn != nil {
		return m.CreateUserFn(ctx, email, firstName, lastName, hashedPassword)
	}
	return m.CreateUserResult, m.CreateUserErr
}

func (m *MockTeamMemberRepository) AddMember(ctx context.Context, orgID, userID uuid.UUID, role string, invitedBy *uuid.UUID) (*entities.TeamMember, error) {
	m.AddMemberCalls++
	if m.AddMemberFn != nil {
		return m.AddMemberFn(ctx, orgID, userID, role, invitedBy)
	}
	return m.AddMemberResult, m.AddMemberErr
}

func (m *MockTeamMemberRepository) UpdateRole(_ context.Context, _ uuid.UUID, _ string) error {
	return m.UpdateRoleErr
}

func (m *MockTeamMemberRepository) Remove(_ context.Context, _ uuid.UUID) error {
	return m.RemoveErr
}

func (m *MockTeamMemberRepository) GetOrganizationIDBySubdomain(ctx context.Context, subdomain string) (uuid.UUID, error) {
	if m.GetOrgIDBySubdomainFn != nil {
		return m.GetOrgIDBySubdomainFn(ctx, subdomain)
	}
	return m.GetOrgIDBySubdomainResult, m.GetOrgIDBySubdomainErr
}
