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
CreatedBy uuid.UUID
CreatedAt time.Time
UpdatedAt time.Time
DeletedAt *time.Time
}

// NewWorkspace creates a new workspace
func NewWorkspace(name, workspaceType string, createdBy uuid.UUID) *Workspace {
return &Workspace{
ID:        uuid.New(),
Name:      name,
Type:      workspaceType,
CreatedBy: createdBy,
CreatedAt: time.Now(),
UpdatedAt: time.Now(),
}
}
