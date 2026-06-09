package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
)

// MockProfileRepository is a hand-rolled mock for repositories.ProfileRepository.
// Tests set the exported fields or Fn callbacks to control behaviour.
type MockProfileRepository struct {
	// Field-based responses (used when the corresponding Fn is nil)
	CreateErr                    error
	GetByIDResult                *entities.SocialProfile
	GetByIDErr                   error
	ListByWorkspaceResult        []*entities.SocialProfile
	ListByWorkspaceErr           error
	UpdateErr                    error
	DeleteErr                    error
	CountActiveByWorkspaceResult int
	CountActiveByWorkspaceErr    error
	ListDueResult                []*entities.SocialProfile
	ListDueErr                   error

	// Function overrides (take precedence when non-nil)
	CreateFn                    func(ctx context.Context, profile *entities.SocialProfile) error
	GetByIDFn                   func(ctx context.Context, id uuid.UUID) (*entities.SocialProfile, error)
	ListByWorkspaceFn           func(ctx context.Context, workspaceID uuid.UUID) ([]*entities.SocialProfile, error)
	UpdateFn                    func(ctx context.Context, profile *entities.SocialProfile) error
	DeleteFn                    func(ctx context.Context, id uuid.UUID) error
	CountActiveByWorkspaceFn    func(ctx context.Context, workspaceID uuid.UUID) (int, error)
	ListDueFn                   func(ctx context.Context, now time.Time, limit int) ([]*entities.SocialProfile, error)

	// Call counters
	CreateCalls                 int
	GetByIDCalls                int
	ListByWorkspaceCalls        int
	UpdateCalls                 int
	DeleteCalls                 int
	CountActiveByWorkspaceCalls int
	ListDueCalls                int
}

func (m *MockProfileRepository) Create(ctx context.Context, profile *entities.SocialProfile) error {
	m.CreateCalls++
	if m.CreateFn != nil {
		return m.CreateFn(ctx, profile)
	}
	return m.CreateErr
}

func (m *MockProfileRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.SocialProfile, error) {
	m.GetByIDCalls++
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return m.GetByIDResult, m.GetByIDErr
}

func (m *MockProfileRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]*entities.SocialProfile, error) {
	m.ListByWorkspaceCalls++
	if m.ListByWorkspaceFn != nil {
		return m.ListByWorkspaceFn(ctx, workspaceID)
	}
	return m.ListByWorkspaceResult, m.ListByWorkspaceErr
}

func (m *MockProfileRepository) Update(ctx context.Context, profile *entities.SocialProfile) error {
	m.UpdateCalls++
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, profile)
	}
	return m.UpdateErr
}

func (m *MockProfileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.DeleteCalls++
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, id)
	}
	return m.DeleteErr
}

func (m *MockProfileRepository) CountActiveByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int, error) {
	m.CountActiveByWorkspaceCalls++
	if m.CountActiveByWorkspaceFn != nil {
		return m.CountActiveByWorkspaceFn(ctx, workspaceID)
	}
	return m.CountActiveByWorkspaceResult, m.CountActiveByWorkspaceErr
}

func (m *MockProfileRepository) ListDue(ctx context.Context, now time.Time, limit int) ([]*entities.SocialProfile, error) {
	m.ListDueCalls++
	if m.ListDueFn != nil {
		return m.ListDueFn(ctx, now, limit)
	}
	return m.ListDueResult, m.ListDueErr
}
