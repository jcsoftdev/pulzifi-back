package list_workspace_members

import "errors"

var (
	ErrWorkspaceNotFound  = errors.New("workspace not found")
	ErrNotWorkspaceMember = errors.New("user is not a member of this workspace")
)
