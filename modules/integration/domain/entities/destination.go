package entities

import (
	"time"

	"github.com/google/uuid"
)

type ScopeType string

const (
	ScopeOrg       ScopeType = "org"
	ScopeWorkspace ScopeType = "workspace"
	ScopePage      ScopeType = "page"
)

type Destination struct {
	ID            uuid.UUID
	IntegrationID *uuid.UUID
	ServiceType   string
	ScopeType     ScopeType
	ScopeID       uuid.UUID
	Target        map[string]any
	Events        []string
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Phase 3 additions — populated only by ListByScope when scope_type is
	// 'workspace' or 'page' (resolved from joined tables for response transport).
	// Empty for org scope.
	WorkspaceName string `json:"workspace_name,omitempty"`
	PageName      string `json:"page_name,omitempty"`
}
