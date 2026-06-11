package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/repositories"
)

// MockChangeRepository is a hand-rolled mock for repositories.ChangeRepository.
type MockChangeRepository struct {
	SaveErr             error
	ListResult          []*entities.SocialChange
	ListErr             error
	GetByIDResult       *entities.SocialChange
	GetByIDErr          error
	UpdateAISummaryErr  error

	SaveFn             func(ctx context.Context, change *entities.SocialChange) error
	ListFn             func(ctx context.Context, filter repositories.ChangeFilter) ([]*entities.SocialChange, error)
	GetByIDFn          func(ctx context.Context, id uuid.UUID) (*entities.SocialChange, error)
	UpdateAISummaryFn  func(ctx context.Context, changeID uuid.UUID, summary string) error

	SaveCalls             int
	ListCalls             int
	GetByIDCalls          int
	UpdateAISummaryCalls  int
}

func (m *MockChangeRepository) Save(ctx context.Context, change *entities.SocialChange) error {
	m.SaveCalls++
	if m.SaveFn != nil {
		return m.SaveFn(ctx, change)
	}
	return m.SaveErr
}

func (m *MockChangeRepository) List(ctx context.Context, filter repositories.ChangeFilter) ([]*entities.SocialChange, error) {
	m.ListCalls++
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return m.ListResult, m.ListErr
}

func (m *MockChangeRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.SocialChange, error) {
	m.GetByIDCalls++
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return m.GetByIDResult, m.GetByIDErr
}

func (m *MockChangeRepository) UpdateAISummary(ctx context.Context, changeID uuid.UUID, summary string) error {
	m.UpdateAISummaryCalls++
	if m.UpdateAISummaryFn != nil {
		return m.UpdateAISummaryFn(ctx, changeID, summary)
	}
	return m.UpdateAISummaryErr
}
