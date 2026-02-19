package update_member_role

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/entities"
	wsmocks "github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/repositories/mocks"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/value_objects"
)

func TestUpdateMemberRoleHandler_Handle(t *testing.T) {
	workspaceID := uuid.New()
	ownerID := uuid.New()
	targetID := uuid.New()

	workspace := &entities.Workspace{
		ID:        workspaceID,
		Name:      "Test Workspace",
		CreatedBy: ownerID,
	}

	ownerMember := &entities.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      ownerID,
		Role:        value_objects.RoleOwner,
	}

	editorMember := &entities.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      ownerID,
		Role:        value_objects.RoleEditor,
	}

	targetMember := &entities.WorkspaceMember{
		WorkspaceID: workspaceID,
		UserID:      targetID,
		Role:        value_objects.RoleViewer,
	}

	tests := []struct {
		name        string
		workspaceID uuid.UUID
		requesterID uuid.UUID
		targetID    uuid.UUID
		req         *UpdateMemberRoleRequest
		setupMock   func(repo *wsmocks.MockWorkspaceRepository)
		wantErr     error
	}{
		{
			name:        "successful role update",
			workspaceID: workspaceID,
			requesterID: ownerID,
			targetID:    targetID,
			req:         &UpdateMemberRoleRequest{Role: "editor"},
			setupMock: func(repo *wsmocks.MockWorkspaceRepository) {
				repo.GetByIDResult = workspace
				repo.GetMemberFn = func(ctx context.Context, wsID, userID uuid.UUID) (*entities.WorkspaceMember, error) {
					if userID == ownerID {
						return ownerMember, nil
					}
					if userID == targetID {
						return targetMember, nil
					}
					return nil, nil
				}
			},
			wantErr: nil,
		},
		{
			name:        "invalid role string",
			workspaceID: workspaceID,
			requesterID: ownerID,
			targetID:    targetID,
			req:         &UpdateMemberRoleRequest{Role: "superadmin"},
			setupMock:   func(repo *wsmocks.MockWorkspaceRepository) {},
			wantErr:     ErrInvalidRole,
		},
		{
			name:        "workspace not found",
			workspaceID: workspaceID,
			requesterID: ownerID,
			targetID:    targetID,
			req:         &UpdateMemberRoleRequest{Role: "editor"},
			setupMock: func(repo *wsmocks.MockWorkspaceRepository) {
				repo.GetByIDResult = nil // workspace does not exist
			},
			wantErr: ErrWorkspaceNotFound,
		},
		{
			name:        "requester is not a workspace member",
			workspaceID: workspaceID,
			requesterID: ownerID,
			targetID:    targetID,
			req:         &UpdateMemberRoleRequest{Role: "editor"},
			setupMock: func(repo *wsmocks.MockWorkspaceRepository) {
				repo.GetByIDResult = workspace
				repo.GetMemberFn = func(ctx context.Context, wsID, userID uuid.UUID) (*entities.WorkspaceMember, error) {
					// requester not found as member
					return nil, nil
				}
			},
			wantErr: ErrNotWorkspaceMember,
		},
		{
			name:        "requester is not an owner (editor cannot manage members)",
			workspaceID: workspaceID,
			requesterID: ownerID,
			targetID:    targetID,
			req:         &UpdateMemberRoleRequest{Role: "editor"},
			setupMock: func(repo *wsmocks.MockWorkspaceRepository) {
				repo.GetByIDResult = workspace
				repo.GetMemberFn = func(ctx context.Context, wsID, userID uuid.UUID) (*entities.WorkspaceMember, error) {
					if userID == ownerID {
						return editorMember, nil // requester is an editor, not owner
					}
					return targetMember, nil
				}
			},
			wantErr: ErrInsufficientPermissions,
		},
		{
			name:        "cannot change own role",
			workspaceID: workspaceID,
			requesterID: ownerID,
			targetID:    ownerID, // same as requester
			req:         &UpdateMemberRoleRequest{Role: "viewer"},
			setupMock: func(repo *wsmocks.MockWorkspaceRepository) {
				repo.GetByIDResult = workspace
				repo.GetMemberFn = func(ctx context.Context, wsID, userID uuid.UUID) (*entities.WorkspaceMember, error) {
					return ownerMember, nil
				}
			},
			wantErr: ErrCannotChangeOwnRole,
		},
		{
			name:        "target user is not a member",
			workspaceID: workspaceID,
			requesterID: ownerID,
			targetID:    uuid.New(), // random user not in workspace
			req:         &UpdateMemberRoleRequest{Role: "editor"},
			setupMock: func(repo *wsmocks.MockWorkspaceRepository) {
				repo.GetByIDResult = workspace
				callCount := 0
				repo.GetMemberFn = func(ctx context.Context, wsID, userID uuid.UUID) (*entities.WorkspaceMember, error) {
					callCount++
					if callCount == 1 {
						// First call: requester lookup
						return ownerMember, nil
					}
					// Second call: target lookup - not found
					return nil, nil
				}
			},
			wantErr: ErrNotWorkspaceMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &wsmocks.MockWorkspaceRepository{}
			if tt.setupMock != nil {
				tt.setupMock(repo)
			}

			handler := NewUpdateMemberRoleHandler(repo)
			err := handler.Handle(context.Background(), tt.workspaceID, tt.requesterID, tt.targetID, tt.req)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if repo.UpdateMemberRoleCalls != 1 {
				t.Errorf("expected 1 UpdateMemberRole call, got %d", repo.UpdateMemberRoleCalls)
			}
		})
	}
}
