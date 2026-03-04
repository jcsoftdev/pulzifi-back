package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/services"
	sharedAI "github.com/jcsoftdev/pulzifi-back/shared/ai"
)

// OpenRouterVisionAnalyzer implements VisionAnalyzer using OpenRouter's multimodal API.
type OpenRouterVisionAnalyzer struct {
	client *sharedAI.OpenRouterClient
}

// NewOpenRouterVisionAnalyzer creates a new vision analyzer with the given OpenRouter client.
func NewOpenRouterVisionAnalyzer(client *sharedAI.OpenRouterClient) *OpenRouterVisionAnalyzer {
	return &OpenRouterVisionAnalyzer{client: client}
}

const visionPrompt = `You are a web page change detection expert. You will be shown two screenshots of the same web page taken at different times.

Analyze both screenshots carefully and determine if there are any meaningful content changes. Focus on:
- Text content changes (prices, numbers, names, descriptions, headlines)
- Added or removed sections, products, listings
- Data changes in tables, charts, or statistics

Ignore these non-meaningful differences:
- Minor styling/CSS differences (colors, fonts, spacing)
- Ad content changes
- Cookie banners appearing/disappearing
- Animation or carousel state differences
- Rendering artifacts or anti-aliasing differences

Respond with a JSON object (no markdown, no code fences):
{
  "has_meaningful_change": true/false,
  "change_summary": "One sentence describing the most important change",
  "change_details": "Detailed description of all meaningful changes found"
}

If no meaningful changes are found, set has_meaningful_change to false and provide an empty change_summary.`

// AnalyzeChange compares two screenshots using vision AI and returns the analysis.
func (v *OpenRouterVisionAnalyzer) AnalyzeChange(ctx context.Context, prevScreenshotB64, currScreenshotB64, pageURL string) (*services.VisionChangeResult, error) {
	messages := []sharedAI.MultimodalMessage{
		{
			Role: "user",
			Content: []sharedAI.ContentBlock{
				{Type: "text", Text: visionPrompt},
				{Type: "text", Text: fmt.Sprintf("Page URL: %s", pageURL)},
				{Type: "text", Text: "BEFORE screenshot:"},
				{Type: "image_url", ImageURL: &sharedAI.ImageURL{URL: "data:image/png;base64," + prevScreenshotB64}},
				{Type: "text", Text: "AFTER screenshot:"},
				{Type: "image_url", ImageURL: &sharedAI.ImageURL{URL: "data:image/png;base64," + currScreenshotB64}},
			},
		},
	}

	response, err := v.client.CompleteMultimodal(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("vision analysis failed: %w", err)
	}

	return parseVisionResponse(response)
}

func parseVisionResponse(response string) (*services.VisionChangeResult, error) {
	// Strip potential markdown code fences
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var result struct {
		HasMeaningfulChange bool   `json:"has_meaningful_change"`
		ChangeSummary       string `json:"change_summary"`
		ChangeDetails       string `json:"change_details"`
	}

	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// If parsing fails, treat the whole response as the summary
		return &services.VisionChangeResult{
			HasMeaningfulChange: true,
			ChangeSummary:       response,
			ChangeDetails:       response,
		}, nil
	}

	return &services.VisionChangeResult{
		HasMeaningfulChange: result.HasMeaningfulChange,
		ChangeSummary:       result.ChangeSummary,
		ChangeDetails:       result.ChangeDetails,
	}, nil
}
