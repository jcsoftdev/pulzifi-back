package mocks

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
)

// MockVisionAnalyzer is a hand-rolled mock for services.VisionAnalyzer.
type MockVisionAnalyzer struct {
	AnalyzeChangeResult *services.VisionChangeResult
	AnalyzeChangeErr    error

	AnalyzeChangeFn func(ctx context.Context, prevScreenshotB64, currScreenshotB64, pageURL string) (*services.VisionChangeResult, error)

	AnalyzeChangeCalls int
}

func (m *MockVisionAnalyzer) AnalyzeChange(ctx context.Context, prevScreenshotB64, currScreenshotB64, pageURL string) (*services.VisionChangeResult, error) {
	m.AnalyzeChangeCalls++
	if m.AnalyzeChangeFn != nil {
		return m.AnalyzeChangeFn(ctx, prevScreenshotB64, currScreenshotB64, pageURL)
	}
	return m.AnalyzeChangeResult, m.AnalyzeChangeErr
}
