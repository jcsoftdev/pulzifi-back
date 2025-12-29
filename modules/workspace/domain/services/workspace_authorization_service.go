package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/value_objects"
)

var (
	ErrMemberNotFound = errors.New("member not found")
)

// WorkspaceAuthorizationService handles workspace-specific authorization logic
// This is DOMAIN-level authorization (Level 2), NOT global system permissions (Level 1)
//
// Authorization Levels:
// - Level 1 (Global): Handled by Auth module - "Can this user use workspaces feature?"
// - Level 2 (Domain): Handled by this service - "What can this user do in THIS workspace?"
type WorkspaceAuthorizationService struct {
	workspaceRepo repositories.WorkspaceRepository
}

func NewWorkspaceAuthorizationService(workspaceRepo repositories.WorkspaceRepository) *WorkspaceAuthorizationService {
	return &WorkspaceAuthorizationService{
		workspaceRepo: workspaceRepo,
	}
}

// CanReadWorkspace checks if user can read the workspace
func (s *WorkspaceAuthorizationService) CanReadWorkspace(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	member, err := s.workspaceRepo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		return false, err
	}

	if member == nil {
		return false, nil
	}

	return member.Role.CanRead(), nil
}

// CanWriteWorkspace checks if user can write to the workspace
func (s *WorkspaceAuthorizationService) CanWriteWorkspace(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	member, err := s.workspaceRepo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		return false, err
	}

	if member == nil {
		return false, nil
	}

	return member.Role.CanWrite(), nil
}

// CanDeleteWorkspace checks if user can delete the workspace
func (s *WorkspaceAuthorizationService) CanDeleteWorkspace(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	member, err := s.workspaceRepo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		return false, err
	}

	if member == nil {
		return false, nil
	}

	return member.Role.CanDelete(), nil
}

// CanManageMembers checks if user can manage workspace members
func (s *WorkspaceAuthorizationService) CanManageMembers(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	member, err := s.workspaceRepo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		return false, err
	}

	if member == nil {
		return false, nil
	}

	return member.Role.CanManageMembers(), nil
}

// GetUserRole returns the user's role in the workspace
func (s *WorkspaceAuthorizationService) GetUserRole(ctx context.Context, workspaceID, userID uuid.UUID) (value_objects.WorkspaceRole, error) {
	member, err := s.workspaceRepo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		return "", err
	}

	if member == nil {
		return "", ErrMemberNotFound
	}

	return member.Role, nil
}

// IsWorkspaceMember checks if user is a member of the workspace
func (s *WorkspaceAuthorizationService) IsWorkspaceMember(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	member, err := s.workspaceRepo.GetMember(ctx, workspaceID, userID)
	if err != nil {
		return false, err
	}

	return member != nil, nil
}
