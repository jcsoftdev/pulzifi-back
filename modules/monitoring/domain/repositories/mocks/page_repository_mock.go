package mocks

import (
	"context"

	"github.com/google/uuid"
)

// MockPageRepository is a hand-rolled test double for the orchestrator.PageRepository interface.
type MockPageRepository struct {
	UpdateLastCheckedErr error
	UpdateLastCheckedFn  func(ctx context.Context, pageID uuid.UUID) error

	UpdateLastCheckedCalls int
}

func (m *MockPageRepository) UpdateLastChecked(ctx context.Context, pageID uuid.UUID) error {
	m.UpdateLastCheckedCalls++
	if m.UpdateLastCheckedFn != nil {
		return m.UpdateLastCheckedFn(ctx, pageID)
	}
	return m.UpdateLastCheckedErr
}
