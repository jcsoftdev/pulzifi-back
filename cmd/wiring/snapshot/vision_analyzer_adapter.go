package snapshotwiring

import (
	"context"

	insightAI "github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/ai"
	snapServices "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
	sharedAI "github.com/jcsoftdev/pulzifi-back/shared/ai"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
)

// visionAnalyzerAdapter implements snapshot's VisionAnalyzer port by wrapping
// insight/infrastructure/ai.OpenRouterVisionAnalyzer.
type visionAnalyzerAdapter struct {
	inner *insightAI.OpenRouterVisionAnalyzer
}

// NewVisionAnalyzer builds a VisionAnalyzer adapter. Returns nil when the
// vision model is not configured.
func NewVisionAnalyzer(cfg *config.Config) snapServices.VisionAnalyzer {
	if cfg.OpenRouterAPIKey == "" || cfg.OpenRouterVisionModel == "" {
		return nil
	}
	visionClient := sharedAI.NewOpenRouterClient(cfg.OpenRouterAPIKey, cfg.OpenRouterVisionModel)
	inner := insightAI.NewOpenRouterVisionAnalyzer(visionClient)
	return &visionAnalyzerAdapter{inner: inner}
}

func (a *visionAnalyzerAdapter) AnalyzeChange(ctx context.Context, prevScreenshotB64, currScreenshotB64, pageURL string) (*snapServices.VisionChangeResult, error) {
	res, err := a.inner.AnalyzeChange(ctx, prevScreenshotB64, currScreenshotB64, pageURL)
	if err != nil {
		return nil, err
	}
	return &snapServices.VisionChangeResult{
		HasMeaningfulChange: res.HasMeaningfulChange,
		ChangeSummary:       res.ChangeSummary,
		ChangeDetails:       res.ChangeDetails,
	}, nil
}
