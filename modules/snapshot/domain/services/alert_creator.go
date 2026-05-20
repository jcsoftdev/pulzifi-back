package services

import (
	"context"

	"github.com/google/uuid"
)

// AlertInput holds the data needed to create a change-detection alert.
type AlertInput struct {
	SchemaName    string
	WorkspaceID   uuid.UUID
	PageID        uuid.UUID
	CheckID       uuid.UUID
	AlertType     string
	Title         string
	Description   string
	ChangeSummary string
}

// AlertCreator is the port for persisting change-detection alerts.
// The concrete adapter wraps alert/infrastructure/persistence.
type AlertCreator interface {
	Create(ctx context.Context, input AlertInput) error
}
