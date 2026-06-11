package mocks

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
)

// MockSocialInsightGenerator is a hand-rolled mock for services.SocialInsightGenerator.
// Generate is safe for concurrent use because run_check calls it from a goroutine.
type MockSocialInsightGenerator struct {
	mu            sync.Mutex
	GenerateErr   error
	GenerateFn    func(ctx context.Context, changeID uuid.UUID, handle, platform string, summary entities.ChangeSummary) error
	GenerateCalls int
}

func (m *MockSocialInsightGenerator) Generate(ctx context.Context, changeID uuid.UUID, handle, platform string, summary entities.ChangeSummary) error {
	m.mu.Lock()
	m.GenerateCalls++
	err := m.GenerateErr
	fn := m.GenerateFn
	m.mu.Unlock()
	if fn != nil {
		return fn(ctx, changeID, handle, platform, summary)
	}
	return err
}

// Calls returns the generate count under lock (use in tests after goroutines settle).
func (m *MockSocialInsightGenerator) Calls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.GenerateCalls
}
