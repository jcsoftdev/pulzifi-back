package mocks

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/entities"
)

type MockInsightGenerator struct {
	GenerateResult []*entities.Insight
	GenerateErr    error

	GenerateFn func(ctx context.Context, pageURL, prevText, newText string, enabledTypes []string) ([]*entities.Insight, error)

	GenerateCalls int
}

func (m *MockInsightGenerator) Generate(ctx context.Context, pageURL, prevText, newText string, enabledTypes []string) ([]*entities.Insight, error) {
	m.GenerateCalls++
	if m.GenerateFn != nil {
		return m.GenerateFn(ctx, pageURL, prevText, newText, enabledTypes)
	}
	return m.GenerateResult, m.GenerateErr
}
