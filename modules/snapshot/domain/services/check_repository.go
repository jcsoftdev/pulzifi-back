package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/entities"
)

// CheckRepository is the port for reading and writing monitoring checks.
// The concrete adapter wraps monitoring/infrastructure/persistence.
type CheckRepository interface {
	// Create persists a new check.
	Create(ctx context.Context, check *entities.Check) error
	// GetByID fetches a check by its ID. Returns nil, nil when not found.
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Check, error)
	// Update persists changes to an existing check.
	Update(ctx context.Context, check *entities.Check) error
	// GetPreviousSuccessfulByPage fetches the most recent successful full-page
	// check for pageID, excluding the check with excludeID.
	GetPreviousSuccessfulByPage(ctx context.Context, pageID, excludeID uuid.UUID) (*entities.Check, error)
	// GetPreviousSuccessfulBySection fetches the most recent successful check
	// for a page+section pair, excluding the check with excludeID.
	GetPreviousSuccessfulBySection(ctx context.Context, pageID uuid.UUID, sectionID *uuid.UUID, excludeID uuid.UUID) (*entities.Check, error)
}
