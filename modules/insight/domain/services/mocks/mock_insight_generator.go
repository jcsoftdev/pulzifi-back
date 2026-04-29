package mocks

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/entities"
)

type MockInsightGenerator struct {
	GenerateResult []*entities.Insight
	GenerateErr    error

	GenerateFn func(ctx context.Context, pageURL, prevText, newText string, enabledTypes []string) ([]*entities.Insight, error)

	GenerateFromDiffResult []*entities.Insight
	GenerateFromDiffErr    error

	GenerateFromDiffFn func(ctx context.Context, pageURL, diffText string, enabledTypes []string) ([]*entities.Insight, error)

	GenerateCalls         int
	GenerateFromDiffCalls int
}

func (m *MockInsightGenerator) Generate(ctx context.Context, pageURL, prevText, newText string, enabledTypes []string) ([]*entities.Insight, error) {
	m.GenerateCalls++
	if m.GenerateFn != nil {
		return m.GenerateFn(ctx, pageURL, prevText, newText, enabledTypes)
	}
	return m.GenerateResult, m.GenerateErr
}

func (m *MockInsightGenerator) GenerateFromDiff(ctx context.Context, pageURL, diffText string, enabledTypes []string) ([]*entities.Insight, error) {
	m.GenerateFromDiffCalls++
	if m.GenerateFromDiffFn != nil {
		return m.GenerateFromDiffFn(ctx, pageURL, diffText, enabledTypes)
	}
	return m.GenerateFromDiffResult, m.GenerateFromDiffErr
}
