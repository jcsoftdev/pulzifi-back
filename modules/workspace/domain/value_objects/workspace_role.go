package value_objects

import (
	"errors"
	"strings"
)

type WorkspaceRole string

const (
	RoleOwner  WorkspaceRole = "owner"
	RoleEditor WorkspaceRole = "editor"
	RoleViewer WorkspaceRole = "viewer"
)

var (
	ErrInvalidRole = errors.New("invalid workspace role")
)

func NewWorkspaceRole(role string) (WorkspaceRole, error) {
	r := WorkspaceRole(strings.ToLower(role))
	switch r {
	case RoleOwner, RoleEditor, RoleViewer:
		return r, nil
	default:
		return "", ErrInvalidRole
	}
}

func (r WorkspaceRole) String() string {
	return string(r)
}

func (r WorkspaceRole) CanRead() bool {
	return r == RoleOwner || r == RoleEditor || r == RoleViewer
}

func (r WorkspaceRole) CanWrite() bool {
	return r == RoleOwner || r == RoleEditor
}

func (r WorkspaceRole) CanDelete() bool {
	return r == RoleOwner
}

func (r WorkspaceRole) CanInvite() bool {
	return r == RoleOwner
}

func (r WorkspaceRole) CanManageMembers() bool {
	return r == RoleOwner
}
