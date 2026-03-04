package services

import "context"

// VisionChangeResult holds the analysis result from vision AI.
type VisionChangeResult struct {
	HasMeaningfulChange bool   // Whether a meaningful content change was detected
	ChangeSummary       string // One-line human-readable summary of the change
	ChangeDetails       string // Detailed description of what changed
}

// VisionAnalyzer analyzes before/after screenshots to detect meaningful changes.
type VisionAnalyzer interface {
	AnalyzeChange(ctx context.Context, prevScreenshotB64, currScreenshotB64, pageURL string) (*VisionChangeResult, error)
}
