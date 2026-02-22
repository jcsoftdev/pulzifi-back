package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/repositories/mocks"
	"github.com/jcsoftdev/pulzifi-back/modules/workspace/domain/value_objects"
)

var (
	testWorkspaceID = uuid.New()
	testUserID      = uuid.New()
	errRepo         = errors.New("repository error")
)

func memberWithRole(role value_objects.WorkspaceRole) *entities.WorkspaceMember {
	return &entities.WorkspaceMember{
		WorkspaceID: testWorkspaceID,
		UserID:      testUserID,
		Role:        role,
	}
}

func TestCanReadWorkspace(t *testing.T) {
	tests := []struct {
		name    string
		member  *entities.WorkspaceMember
		repoErr error
		want    bool
		wantErr bool
	}{
		{"owner can read", memberWithRole(value_objects.RoleOwner), nil, true, false},
		{"editor can read", memberWithRole(value_objects.RoleEditor), nil, true, false},
		{"viewer can read", memberWithRole(value_objects.RoleViewer), nil, true, false},
		{"non-member cannot read", nil, nil, false, false},
		{"repo error propagates", nil, errRepo, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockWorkspaceRepository{
				GetMemberResult: tt.member,
				GetMemberErr:    tt.repoErr,
			}
			svc := NewWorkspaceAuthorizationService(repo)
			got, err := svc.CanReadWorkspace(context.Background(), testWorkspaceID, testUserID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if got != tt.want {
				t.Fatalf("want=%v, got=%v", tt.want, got)
			}
		})
	}
}

func TestCanWriteWorkspace(t *testing.T) {
	tests := []struct {
		name    string
		member  *entities.WorkspaceMember
		repoErr error
		want    bool
		wantErr bool
	}{
		{"owner can write", memberWithRole(value_objects.RoleOwner), nil, true, false},
		{"editor can write", memberWithRole(value_objects.RoleEditor), nil, true, false},
		{"viewer cannot write", memberWithRole(value_objects.RoleViewer), nil, false, false},
		{"non-member cannot write", nil, nil, false, false},
		{"repo error propagates", nil, errRepo, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockWorkspaceRepository{
				GetMemberResult: tt.member,
				GetMemberErr:    tt.repoErr,
			}
			svc := NewWorkspaceAuthorizationService(repo)
			got, err := svc.CanWriteWorkspace(context.Background(), testWorkspaceID, testUserID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if got != tt.want {
				t.Fatalf("want=%v, got=%v", tt.want, got)
			}
		})
	}
}

func TestCanDeleteWorkspace(t *testing.T) {
	tests := []struct {
		name    string
		member  *entities.WorkspaceMember
		repoErr error
		want    bool
		wantErr bool
	}{
		{"owner can delete", memberWithRole(value_objects.RoleOwner), nil, true, false},
		{"editor cannot delete", memberWithRole(value_objects.RoleEditor), nil, false, false},
		{"viewer cannot delete", memberWithRole(value_objects.RoleViewer), nil, false, false},
		{"non-member cannot delete", nil, nil, false, false},
		{"repo error propagates", nil, errRepo, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockWorkspaceRepository{
				GetMemberResult: tt.member,
				GetMemberErr:    tt.repoErr,
			}
			svc := NewWorkspaceAuthorizationService(repo)
			got, err := svc.CanDeleteWorkspace(context.Background(), testWorkspaceID, testUserID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if got != tt.want {
				t.Fatalf("want=%v, got=%v", tt.want, got)
			}
		})
	}
}

func TestCanManageMembers(t *testing.T) {
	tests := []struct {
		name    string
		member  *entities.WorkspaceMember
		repoErr error
		want    bool
		wantErr bool
	}{
		{"owner can manage", memberWithRole(value_objects.RoleOwner), nil, true, false},
		{"editor cannot manage", memberWithRole(value_objects.RoleEditor), nil, false, false},
		{"viewer cannot manage", memberWithRole(value_objects.RoleViewer), nil, false, false},
		{"non-member cannot manage", nil, nil, false, false},
		{"repo error propagates", nil, errRepo, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockWorkspaceRepository{
				GetMemberResult: tt.member,
				GetMemberErr:    tt.repoErr,
			}
			svc := NewWorkspaceAuthorizationService(repo)
			got, err := svc.CanManageMembers(context.Background(), testWorkspaceID, testUserID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if got != tt.want {
				t.Fatalf("want=%v, got=%v", tt.want, got)
			}
		})
	}
}

func TestGetUserRole(t *testing.T) {
	tests := []struct {
		name     string
		member   *entities.WorkspaceMember
		repoErr  error
		wantRole value_objects.WorkspaceRole
		wantErr  bool
	}{
		{"returns owner role", memberWithRole(value_objects.RoleOwner), nil, value_objects.RoleOwner, false},
		{"returns editor role", memberWithRole(value_objects.RoleEditor), nil, value_objects.RoleEditor, false},
		{"non-member returns error", nil, nil, "", true},
		{"repo error propagates", nil, errRepo, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockWorkspaceRepository{
				GetMemberResult: tt.member,
				GetMemberErr:    tt.repoErr,
			}
			svc := NewWorkspaceAuthorizationService(repo)
			role, err := svc.GetUserRole(context.Background(), testWorkspaceID, testUserID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if role != tt.wantRole {
				t.Fatalf("want role=%v, got=%v", tt.wantRole, role)
			}
		})
	}
}

func TestIsWorkspaceMember(t *testing.T) {
	tests := []struct {
		name    string
		member  *entities.WorkspaceMember
		repoErr error
		want    bool
		wantErr bool
	}{
		{"is member", memberWithRole(value_objects.RoleViewer), nil, true, false},
		{"not member", nil, nil, false, false},
		{"repo error propagates", nil, errRepo, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mocks.MockWorkspaceRepository{
				GetMemberResult: tt.member,
				GetMemberErr:    tt.repoErr,
			}
			svc := NewWorkspaceAuthorizationService(repo)
			got, err := svc.IsWorkspaceMember(context.Background(), testWorkspaceID, testUserID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
			if got != tt.want {
				t.Fatalf("want=%v, got=%v", tt.want, got)
			}
		})
	}
}
