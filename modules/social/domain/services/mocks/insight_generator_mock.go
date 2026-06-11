package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/entities"
)

// MockSocialInsightGenerator is a hand-rolled mock for services.SocialInsightGenerator.
type MockSocialInsightGenerator struct {
	GenerateErr   error
	GenerateFn    func(ctx context.Context, changeID uuid.UUID, handle, platform string, summary entities.ChangeSummary) error
	GenerateCalls int
}

func (m *MockSocialInsightGenerator) Generate(ctx context.Context, changeID uuid.UUID, handle, platform string, summary entities.ChangeSummary) error {
	m.GenerateCalls++
	if m.GenerateFn != nil {
		return m.GenerateFn(ctx, changeID, handle, platform, summary)
	}
	return m.GenerateErr
}
