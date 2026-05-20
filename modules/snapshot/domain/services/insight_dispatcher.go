package services

import (
	"context"

	"github.com/google/uuid"
)

// InsightRequest holds the data needed to trigger AI insight generation.
type InsightRequest struct {
	PageID              uuid.UUID
	CheckID             uuid.UUID
	PageURL             string
	PrevText            string
	NewText             string
	DiffText            string
	SchemaName          string
	EnabledInsightTypes []string
}

// InsightDispatcher is the port for triggering AI insight generation after a change.
// The concrete adapter wraps insight/application/generate_insights.Handler.
// Returns nil when insights are disabled (e.g. no API key configured).
type InsightDispatcher interface {
	// Dispatch generates insights for a detected page change. May be a no-op
	// when the dispatcher is not configured.
	Dispatch(ctx context.Context, req InsightRequest) error
}
