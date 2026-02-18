package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sharedAI "github.com/jcsoftdev/pulzifi-back/shared/ai"

	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/entities"
)

const maxTextLen = 3000

// insightTypeDescriptions maps a type ID to a human-readable description for the prompt.
var insightTypeDescriptions = map[string]string{
	"marketing":              "Marketing Lens: analyze the messaging and positioning changes from a marketing perspective",
	"market_analysis":        "Market Analysis: identify market opportunities and competitive implications of these changes",
	"business_opportunities": "Business Opportunities: surface specific business opportunities arising from these changes",
	"job_recommendation":     "Job Recommendation: identify hiring or talent trends visible from these changes",
	"overview":               "Overview: provide a 2-3 paragraph summary of what changed and its strategic significance",
}

type insightPayload struct {
	InsightType string `json:"insight_type"`
	Title       string `json:"title"`
	Content     string `json:"content"`
}

// OpenRouterGenerator implements InsightGenerator using the OpenRouter API.
type OpenRouterGenerator struct {
	client *sharedAI.OpenRouterClient
}

// NewOpenRouterGenerator creates a new OpenRouterGenerator.
func NewOpenRouterGenerator(client *sharedAI.OpenRouterClient) *OpenRouterGenerator {
	return &OpenRouterGenerator{client: client}
}

// Generate calls the LLM to produce insights for the requested types only.
func (g *OpenRouterGenerator) Generate(ctx context.Context, pageURL, prevText, newText string, enabledTypes []string) ([]*entities.Insight, error) {
	if len(enabledTypes) == 0 {
		return nil, nil
	}

	prevText = truncate(prevText, maxTextLen)
	newText = truncate(newText, maxTextLen)

	// Build the JSON template and description list for the requested types
	var descriptions []string
	var jsonTemplate []string
	for _, t := range enabledTypes {
		desc, ok := insightTypeDescriptions[t]
		if !ok {
			desc = fmt.Sprintf("%s: analyze this aspect of the changes", t)
		}
		descriptions = append(descriptions, fmt.Sprintf("- %s", desc))
		jsonTemplate = append(jsonTemplate, fmt.Sprintf(`  {"insight_type": %q, "title": "<title>", "content": "<content>"}`, t))
	}

	prompt := fmt.Sprintf(`You are a competitive intelligence analyst. A monitoring system detected content changes on the webpage: %s

Analyze the following changes and generate exactly %d insight(s) covering these perspectives:
%s

PREVIOUS TEXT:
---
%s
---

NEW TEXT:
+++
%s
+++

Return ONLY a valid JSON array with exactly %d object(s). No markdown, no code blocks, no explanation — pure JSON:
[
%s
]`,
		pageURL,
		len(enabledTypes),
		strings.Join(descriptions, "\n"),
		prevText,
		newText,
		len(enabledTypes),
		strings.Join(jsonTemplate, ",\n"),
	)

	messages := []sharedAI.Message{
		{Role: "user", Content: prompt},
	}

	raw, err := g.client.Complete(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("openrouter generator: %w", err)
	}

	payloads, err := parseInsights(raw)
	if err != nil {
		return nil, fmt.Errorf("openrouter generator: parse response: %w", err)
	}

	insights := make([]*entities.Insight, 0, len(payloads))
	for _, p := range payloads {
		insights = append(insights, &entities.Insight{
			InsightType: p.InsightType,
			Title:       p.Title,
			Content:     p.Content,
		})
	}
	return insights, nil
}

// parseInsights extracts a JSON array from the raw LLM output,
// handling both plain JSON and markdown-fenced responses.
func parseInsights(raw string) ([]insightPayload, error) {
	raw = strings.TrimSpace(raw)

	// Strip markdown code fences if present (```json ... ``` or ``` ... ```)
	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimPrefix(raw, "```")
		if idx := strings.LastIndex(raw, "```"); idx != -1 {
			raw = raw[:idx]
		}
		raw = strings.TrimSpace(raw)
	}

	// Find the outermost JSON array
	start := strings.Index(raw, "[")
	end := strings.LastIndex(raw, "]")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON array found in LLM response")
	}
	raw = raw[start : end+1]

	var payloads []insightPayload
	if err := json.Unmarshal([]byte(raw), &payloads); err != nil {
		return nil, fmt.Errorf("unmarshal insights: %w", err)
	}
	return payloads, nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "... [truncated]"
}
