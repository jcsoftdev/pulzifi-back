package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/entities"
)

// MonitoringReader is the port for reading monitoring configuration and sections.
// The concrete adapter wraps monitoring/infrastructure/persistence.
type MonitoringReader interface {
	// GetConfigByPageID returns the monitoring config for a page, or nil if not found.
	GetConfigByPageID(ctx context.Context, schemaName string, pageID uuid.UUID) (*entities.MonitoringConfig, error)
	// ListSectionsByPageID returns all monitored sections for a page.
	ListSectionsByPageID(ctx context.Context, schemaName string, pageID uuid.UUID) ([]*entities.MonitoredSection, error)
	// GetEmailEnabledByPage returns all notification preferences for a page
	// that have email notifications enabled.
	GetEmailEnabledByPage(ctx context.Context, schemaName string, pageID uuid.UUID) ([]*entities.NotificationPreference, error)
	// GetUserEmail returns the email address for a user by ID.
	GetUserEmail(ctx context.Context, userID uuid.UUID) (string, error)
	// GetWorkspaceIDForPage returns the workspace_id for a given page.
	GetWorkspaceIDForPage(ctx context.Context, schemaName string, pageID uuid.UUID) (uuid.UUID, error)
	// GetPageTitle returns the title for a given page.
	GetPageTitle(ctx context.Context, schemaName string, pageID uuid.UUID) (string, error)
	// GetOrgIDBySchemaName returns the org ID for a given tenant schema name.
	GetOrgIDBySchemaName(ctx context.Context, schemaName string) (uuid.UUID, error)
	// UpdatePageSnapshotMetadata updates the thumbnail_url and last_change_detected_at for a page.
	UpdatePageSnapshotMetadata(ctx context.Context, schemaName string, pageID uuid.UUID, thumbnailURL string, changeDetected bool) error
}
