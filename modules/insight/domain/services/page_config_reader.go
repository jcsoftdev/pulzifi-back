package services

import (
	"context"

	"github.com/google/uuid"
)

// PageInsightConfig holds the insight-relevant fields of a monitoring config.
type PageInsightConfig struct {
	EnabledInsightTypes []string
}

// PageConfigReader abstracts fetching monitoring config data needed for insight generation.
// It lives in insight's domain so the insight module has no import dependency on the
// monitoring module.
type PageConfigReader interface {
	// GetInsightConfig returns the insight-related config for a page, or nil if none exists.
	GetInsightConfig(ctx context.Context, pageID uuid.UUID) (*PageInsightConfig, error)
	// GetPageURL returns the URL of the monitored page.
	GetPageURL(ctx context.Context, pageID uuid.UUID) (string, error)
}

// CheckSummary holds the insight-relevant fields of a monitoring check.
type CheckSummary struct {
	ID              uuid.UUID
	Status          string
	HTMLSnapshotURL string
}

// CheckReader abstracts fetching check data needed for insight generation.
// It lives in insight's domain so the insight module has no import dependency on the
// monitoring module.
type CheckReader interface {
	// GetByID returns a single check by ID, or nil if not found.
	GetByID(ctx context.Context, checkID uuid.UUID) (*CheckSummary, error)
	// ListByPage returns all checks for the given page.
	ListByPage(ctx context.Context, pageID uuid.UUID) ([]*CheckSummary, error)
}
