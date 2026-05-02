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
}
