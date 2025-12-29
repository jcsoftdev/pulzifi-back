package delete_workspace

import "errors"

var (
	// ErrWorkspaceNotFound is returned when workspace doesn't exist
	ErrWorkspaceNotFound = errors.New("workspace not found")

	// ErrWorkspaceNotOwned is returned when user doesn't own the workspace
	ErrWorkspaceNotOwned = errors.New("workspace not owned by user")
)
