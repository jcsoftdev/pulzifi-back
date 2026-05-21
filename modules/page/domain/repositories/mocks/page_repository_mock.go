package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/page/domain/entities"
)

// MockPageRepository is a hand-rolled mock for repositories.PageRepository.
type MockPageRepository struct {
	CreateErr          error
	GetByIDResult      *entities.Page
	GetByIDErr         error
	ListByWorkResult   []*entities.Page
	ListByWorkErr      error
	UpdateErr          error
	DeleteErr          error
	BulkDeleteErr      error

	CreateFn       func(ctx context.Context, page *entities.Page) error
	GetByIDFn      func(ctx context.Context, id uuid.UUID) (*entities.Page, error)
	ListByWorkFn   func(ctx context.Context, workspaceID uuid.UUID) ([]*entities.Page, error)

	CreateCalls      int
	GetByIDCalls     int
	ListByWorkCalls  int
}

func (m *MockPageRepository) Create(ctx context.Context, page *entities.Page) error {
	m.CreateCalls++
	if m.CreateFn != nil {
		return m.CreateFn(ctx, page)
	}
	return m.CreateErr
}

func (m *MockPageRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Page, error) {
	m.GetByIDCalls++
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return m.GetByIDResult, m.GetByIDErr
}

func (m *MockPageRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entities.Page, error) {
	m.ListByWorkCalls++
	if m.ListByWorkFn != nil {
		return m.ListByWorkFn(ctx, workspaceID)
	}
	return m.ListByWorkResult, m.ListByWorkErr
}

func (m *MockPageRepository) Update(_ context.Context, _ *entities.Page) error {
	return m.UpdateErr
}

func (m *MockPageRepository) Delete(_ context.Context, _ uuid.UUID) error {
	return m.DeleteErr
}

func (m *MockPageRepository) BulkDelete(_ context.Context, _ []uuid.UUID) error {
	return m.BulkDeleteErr
}
