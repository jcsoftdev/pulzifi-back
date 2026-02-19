package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/value_objects"
)

// MockWorkspaceRepository is a mock implementation of repositories.WorkspaceRepository.
type MockWorkspaceRepository struct {
	// Default return values
	CreateErr         error
	GetByIDResult     *entities.Workspace
	GetByIDErr        error
	ListResult        []*entities.Workspace
	ListErr           error
	ListByCreatorResult []*entities.Workspace
	ListByCreatorErr  error
	UpdateErr         error
	DeleteErr         error

	AddMemberErr         error
	GetMemberResult      *entities.WorkspaceMember
	GetMemberErr         error
	ListMembersResult    []*entities.WorkspaceMember
	ListMembersErr       error
	ListByMemberResult   []*entities.Workspace
	ListByMemberErr      error
	UpdateMemberRoleErr  error
	RemoveMemberErr      error

	// Function hooks
	CreateFn           func(ctx context.Context, workspace *entities.Workspace) error
	GetByIDFn          func(ctx context.Context, id uuid.UUID) (*entities.Workspace, error)
	ListFn             func(ctx context.Context) ([]*entities.Workspace, error)
	ListByCreatorFn    func(ctx context.Context, createdBy uuid.UUID) ([]*entities.Workspace, error)
	UpdateFn           func(ctx context.Context, workspace *entities.Workspace) error
	DeleteFn           func(ctx context.Context, id uuid.UUID) error
	AddMemberFn        func(ctx context.Context, member *entities.WorkspaceMember) error
	GetMemberFn        func(ctx context.Context, workspaceID, userID uuid.UUID) (*entities.WorkspaceMember, error)
	ListMembersFn      func(ctx context.Context, workspaceID uuid.UUID) ([]*entities.WorkspaceMember, error)
	ListByMemberFn     func(ctx context.Context, userID uuid.UUID) ([]*entities.Workspace, error)
	UpdateMemberRoleFn func(ctx context.Context, workspaceID, userID uuid.UUID, role value_objects.WorkspaceRole) error
	RemoveMemberFn     func(ctx context.Context, workspaceID, userID uuid.UUID) error

	// Call tracking
	GetByIDCalls        int
	GetMemberCalls      int
	UpdateMemberRoleCalls int
}

func (m *MockWorkspaceRepository) Create(ctx context.Context, workspace *entities.Workspace) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, workspace)
	}
	return m.CreateErr
}

func (m *MockWorkspaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Workspace, error) {
	m.GetByIDCalls++
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return m.GetByIDResult, m.GetByIDErr
}

func (m *MockWorkspaceRepository) List(ctx context.Context) ([]*entities.Workspace, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx)
	}
	return m.ListResult, m.ListErr
}

func (m *MockWorkspaceRepository) ListByCreator(ctx context.Context, createdBy uuid.UUID) ([]*entities.Workspace, error) {
	if m.ListByCreatorFn != nil {
		return m.ListByCreatorFn(ctx, createdBy)
	}
	return m.ListByCreatorResult, m.ListByCreatorErr
}

func (m *MockWorkspaceRepository) Update(ctx context.Context, workspace *entities.Workspace) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, workspace)
	}
	return m.UpdateErr
}

func (m *MockWorkspaceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return m.DeleteErr
}

func (m *MockWorkspaceRepository) AddMember(ctx context.Context, member *entities.WorkspaceMember) error {
	if m.AddMemberFn != nil {
		return m.AddMemberFn(ctx, member)
	}
	return m.AddMemberErr
}

func (m *MockWorkspaceRepository) GetMember(ctx context.Context, workspaceID, userID uuid.UUID) (*entities.WorkspaceMember, error) {
	m.GetMemberCalls++
	if m.GetMemberFn != nil {
		return m.GetMemberFn(ctx, workspaceID, userID)
	}
	return m.GetMemberResult, m.GetMemberErr
}

func (m *MockWorkspaceRepository) ListMembers(ctx context.Context, workspaceID uuid.UUID) ([]*entities.WorkspaceMember, error) {
	if m.ListMembersFn != nil {
		return m.ListMembersFn(ctx, workspaceID)
	}
	return m.ListMembersResult, m.ListMembersErr
}

func (m *MockWorkspaceRepository) ListByMember(ctx context.Context, userID uuid.UUID) ([]*entities.Workspace, error) {
	if m.ListByMemberFn != nil {
		return m.ListByMemberFn(ctx, userID)
	}
	return m.ListByMemberResult, m.ListByMemberErr
}

func (m *MockWorkspaceRepository) UpdateMemberRole(ctx context.Context, workspaceID, userID uuid.UUID, role value_objects.WorkspaceRole) error {
	m.UpdateMemberRoleCalls++
	if m.UpdateMemberRoleFn != nil {
		return m.UpdateMemberRoleFn(ctx, workspaceID, userID, role)
	}
	return m.UpdateMemberRoleErr
}

func (m *MockWorkspaceRepository) RemoveMember(ctx context.Context, workspaceID, userID uuid.UUID) error {
	if m.RemoveMemberFn != nil {
		return m.RemoveMemberFn(ctx, workspaceID, userID)
	}
	return m.RemoveMemberErr
}
