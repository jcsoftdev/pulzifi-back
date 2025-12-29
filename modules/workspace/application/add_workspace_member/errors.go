package add_workspace_member

import "errors"

var (
	ErrWorkspaceNotFound       = errors.New("workspace not found")
	ErrNotWorkspaceMember      = errors.New("user is not a member of this workspace")
	ErrInsufficientPermissions = errors.New("user does not have permission to invite members")
	ErrMemberAlreadyExists     = errors.New("user is already a member of this workspace")
)
