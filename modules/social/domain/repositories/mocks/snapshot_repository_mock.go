package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
)

// MockSnapshotRepository is a hand-rolled mock for repositories.SnapshotRepository.
type MockSnapshotRepository struct {
	SaveErr                    error
	GetLatestByProfileResult   *entities.SocialSnapshot
	GetLatestByProfileErr      error
	GetByIDResult              *entities.SocialSnapshot
	GetByIDErr                 error

	SaveFn                    func(ctx context.Context, snapshot *entities.SocialSnapshot) error
	GetLatestByProfileFn      func(ctx context.Context, profileID uuid.UUID) (*entities.SocialSnapshot, error)
	GetByIDFn                 func(ctx context.Context, id uuid.UUID) (*entities.SocialSnapshot, error)

	SaveCalls               int
	GetLatestByProfileCalls int
	GetByIDCalls            int
}

func (m *MockSnapshotRepository) Save(ctx context.Context, snapshot *entities.SocialSnapshot) error {
	m.SaveCalls++
	if m.SaveFn != nil {
		return m.SaveFn(ctx, snapshot)
	}
	return m.SaveErr
}

func (m *MockSnapshotRepository) GetLatestByProfile(ctx context.Context, profileID uuid.UUID) (*entities.SocialSnapshot, error) {
	m.GetLatestByProfileCalls++
	if m.GetLatestByProfileFn != nil {
		return m.GetLatestByProfileFn(ctx, profileID)
	}
	return m.GetLatestByProfileResult, m.GetLatestByProfileErr
}

func (m *MockSnapshotRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.SocialSnapshot, error) {
	m.GetByIDCalls++
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return m.GetByIDResult, m.GetByIDErr
}
