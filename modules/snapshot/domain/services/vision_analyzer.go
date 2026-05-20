package services

import "context"

// VisionChangeResult holds the result of vision AI change analysis.
type VisionChangeResult struct {
	HasMeaningfulChange bool
	ChangeSummary       string
	ChangeDetails       string
}

// VisionAnalyzer is the port for AI-powered screenshot comparison.
// The concrete adapter wraps insight/infrastructure/ai.OpenRouterVisionAnalyzer.
type VisionAnalyzer interface {
	AnalyzeChange(ctx context.Context, prevScreenshotB64, currScreenshotB64, pageURL string) (*VisionChangeResult, error)
}
