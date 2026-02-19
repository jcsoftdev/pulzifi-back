package update_member_role

import "errors"

var (
	ErrWorkspaceNotFound       = errors.New("workspace not found")
	ErrNotWorkspaceMember      = errors.New("user is not a member of this workspace")
	ErrInsufficientPermissions = errors.New("user does not have permission to update member roles")
	ErrInvalidRole             = errors.New("invalid workspace role")
	ErrCannotChangeOwnRole     = errors.New("cannot change your own role")
)
