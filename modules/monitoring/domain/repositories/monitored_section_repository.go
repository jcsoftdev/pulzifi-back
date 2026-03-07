package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
)

// MonitoredSectionRepository defines operations for managing monitored sections.
type MonitoredSectionRepository interface {
	Create(ctx context.Context, section *entities.MonitoredSection) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.MonitoredSection, error)
	ListByPageID(ctx context.Context, pageID uuid.UUID) ([]*entities.MonitoredSection, error)
	Update(ctx context.Context, section *entities.MonitoredSection) error
	Delete(ctx context.Context, id uuid.UUID) error
	// ReplaceAll atomically replaces all sections for a page.
	ReplaceAll(ctx context.Context, pageID uuid.UUID, sections []*entities.MonitoredSection) error
}
