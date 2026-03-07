package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/monitoring/domain/entities"
)

// CheckRepository defines operations for managing checks
type CheckRepository interface {
	Create(ctx context.Context, check *entities.Check) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Check, error)
	ListByPage(ctx context.Context, pageID uuid.UUID) ([]*entities.Check, error)
	// ListByPageAndSection returns checks for a page filtered by section. If sectionID is nil, returns only full-page checks.
	ListByPageAndSection(ctx context.Context, pageID uuid.UUID, sectionID *uuid.UUID) ([]*entities.Check, error)
	// ListByParentCheckID returns all section checks that belong to a parent check.
	ListByParentCheckID(ctx context.Context, parentCheckID uuid.UUID) ([]*entities.Check, error)
	// ListSectionChecksByPage returns all section checks for a page (section_id IS NOT NULL).
	ListSectionChecksByPage(ctx context.Context, pageID uuid.UUID) ([]*entities.Check, error)
	GetLatestByPage(ctx context.Context, pageID uuid.UUID) (*entities.Check, error)
	Update(ctx context.Context, check *entities.Check) error
	GetPreviousSuccessfulByPage(ctx context.Context, pageID, excludeCheckID uuid.UUID) (*entities.Check, error)
	// GetPreviousSuccessfulBySection returns the most recent successful check for the same section.
	GetPreviousSuccessfulBySection(ctx context.Context, pageID uuid.UUID, sectionID *uuid.UUID, excludeCheckID uuid.UUID) (*entities.Check, error)
}
