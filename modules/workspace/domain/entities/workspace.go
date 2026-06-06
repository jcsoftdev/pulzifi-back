package entities

import (
"time"
"github.com/google/uuid"
)

// Workspace represents a workspace
type Workspace struct {
ID        uuid.UUID
Name      string
	Type      string
	Tags      []string
	CreatedBy uuid.UUID
	CreatedAt time.Time
UpdatedAt time.Time
DeletedAt *time.Time
// PageCount is the number of non-deleted pages in the workspace. It is only
// populated by List (the workspace-grid query); other reads leave it 0.
PageCount int
}

// NewWorkspace creates a new workspace
func NewWorkspace(name, workspaceType string, tags []string, createdBy uuid.UUID) *Workspace {
	return &Workspace{
		ID:        uuid.New(),
		Name:      name,
		Type:      workspaceType,
		Tags:      tags,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
