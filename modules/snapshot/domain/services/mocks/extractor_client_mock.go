package mocks

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
)

// MockExtractorClient is a hand-rolled mock for services.ExtractorClient.
type MockExtractorClient struct {
	ExtractResult *services.ExtractorResult
	ExtractErr    error

	ExtractFn func(ctx context.Context, url string, opts services.ExtractOptions) (*services.ExtractorResult, error)

	ExtractCalls int
}

func (m *MockExtractorClient) Extract(ctx context.Context, url string, opts services.ExtractOptions) (*services.ExtractorResult, error) {
	m.ExtractCalls++
	if m.ExtractFn != nil {
		return m.ExtractFn(ctx, url, opts)
	}
	return m.ExtractResult, m.ExtractErr
}
