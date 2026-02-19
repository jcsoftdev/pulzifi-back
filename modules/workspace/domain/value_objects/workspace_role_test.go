package value_objects

import (
	"errors"
	"testing"
)

func TestNewWorkspaceRole(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    WorkspaceRole
		wantErr bool
	}{
		{"valid owner lowercase", "owner", RoleOwner, false},
		{"valid editor lowercase", "editor", RoleEditor, false},
		{"valid viewer lowercase", "viewer", RoleViewer, false},
		{"valid owner uppercase", "OWNER", RoleOwner, false},
		{"valid editor mixed case", "Editor", RoleEditor, false},
		{"valid viewer mixed case", "Viewer", RoleViewer, false},
		{"invalid role admin", "admin", "", true},
		{"invalid role empty", "", "", true},
		{"invalid role random", "superuser", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWorkspaceRole(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewWorkspaceRole(%q) expected error, got nil", tt.input)
				}
				if !errors.Is(err, ErrInvalidRole) {
					t.Errorf("NewWorkspaceRole(%q) error = %v, want ErrInvalidRole", tt.input, err)
				}
				return
			}
			if err != nil {
				t.Errorf("NewWorkspaceRole(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("NewWorkspaceRole(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestWorkspaceRoleString(t *testing.T) {
	tests := []struct {
		role WorkspaceRole
		want string
	}{
		{RoleOwner, "owner"},
		{RoleEditor, "editor"},
		{RoleViewer, "viewer"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.role.String(); got != tt.want {
				t.Errorf("WorkspaceRole.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWorkspaceRolePermissions(t *testing.T) {
	tests := []struct {
		name             string
		role             WorkspaceRole
		canRead          bool
		canWrite         bool
		canDelete        bool
		canInvite        bool
		canManageMembers bool
	}{
		{
			name:             "owner has all permissions",
			role:             RoleOwner,
			canRead:          true,
			canWrite:         true,
			canDelete:        true,
			canInvite:        true,
			canManageMembers: true,
		},
		{
			name:             "editor can read and write only",
			role:             RoleEditor,
			canRead:          true,
			canWrite:         true,
			canDelete:        false,
			canInvite:        false,
			canManageMembers: false,
		},
		{
			name:             "viewer can read only",
			role:             RoleViewer,
			canRead:          true,
			canWrite:         false,
			canDelete:        false,
			canInvite:        false,
			canManageMembers: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.CanRead(); got != tt.canRead {
				t.Errorf("%s.CanRead() = %v, want %v", tt.role, got, tt.canRead)
			}
			if got := tt.role.CanWrite(); got != tt.canWrite {
				t.Errorf("%s.CanWrite() = %v, want %v", tt.role, got, tt.canWrite)
			}
			if got := tt.role.CanDelete(); got != tt.canDelete {
				t.Errorf("%s.CanDelete() = %v, want %v", tt.role, got, tt.canDelete)
			}
			if got := tt.role.CanInvite(); got != tt.canInvite {
				t.Errorf("%s.CanInvite() = %v, want %v", tt.role, got, tt.canInvite)
			}
			if got := tt.role.CanManageMembers(); got != tt.canManageMembers {
				t.Errorf("%s.CanManageMembers() = %v, want %v", tt.role, got, tt.canManageMembers)
			}
		})
	}
}
