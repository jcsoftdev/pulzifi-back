package remove_workspace_member

import "errors"

var (
	ErrWorkspaceNotFound       = errors.New("workspace not found")
	ErrNotWorkspaceMember      = errors.New("user is not a member of this workspace")
	ErrInsufficientPermissions = errors.New("user does not have permission to remove members")
	ErrCannotRemoveOwner       = errors.New("cannot remove the owner from workspace")
	ErrCannotRemoveSelf        = errors.New("cannot remove yourself from workspace")
	ErrMemberNotFound          = errors.New("member not found in workspace")
)
