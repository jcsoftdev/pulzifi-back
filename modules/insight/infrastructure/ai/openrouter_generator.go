package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sharedAI "github.com/jcsoftdev/pulzifi-back/shared/ai"

	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/entities"
)

const maxTextLen = 4000

// insightTypeDescriptions maps a type ID to a detailed prompt instruction.
var insightTypeDescriptions = map[string]string{
	"overview": `Overview — Write a concise 2–3 paragraph strategic summary of what changed and why it matters.
  • Paragraph 1: Describe the core shift in messaging, content, or positioning in concrete terms (reference specific changes).
  • Paragraph 2: Explain the broader significance — what does this signal about the company's strategy or direction?
  • Keep it factual, sharp, and free of filler words.`,

	"marketing": `Marketing Lens — Analyse the messaging and positioning changes from a marketing perspective across 3 sub-sections:
  1. Brand repositioning: How is the brand changing how it presents itself?
  2. Messaging emphasis: What themes, promises, or emotions are now foregrounded? What was de-emphasised?
  3. Likely downstream implications: How might this change campaigns, channels, or target audiences?
  Write 1–2 paragraphs per sub-section.`,

	"market_analysis": `Market Analysis — Identify market opportunities and competitive implications using numbered points:
  For each point use the format:
    [number]. [Bold sub-heading]
    [1–2 paragraph explanation with concrete reasoning]
  Cover at least: market segment shifts, competitive positioning, and differentiation opportunities.`,

	"business_opportunities": `Business Opportunities — Surface 3–5 specific, actionable opportunities that arise from these changes.
  For each opportunity:
    • State the opportunity in one sentence.
    • Explain the reasoning (why does this change create this opening?).
    • Suggest a concrete action a competitor or partner could take.`,

	"job_recommendation": `Job & Talent Signals — Infer hiring and organisational priorities from the content changes.
  • What capabilities or functions are being amplified?
  • What roles might this company be investing in, or what skill sets are becoming strategically important?
  • What does this signal about internal priorities or team structure?`,
}

// insightTitles provides the display title for each type.
var insightTitles = map[string]string{
	"overview":               "Overview",
	"marketing":              "Marketing Lens",
	"market_analysis":        "Market Analysis",
	"business_opportunities": "Business Opportunities",
	"job_recommendation":     "Job & Talent Signals",
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

// Generate calls the LLM to produce insights. Overview is always prepended first,
// followed by the user-enabled types in order.
func (g *OpenRouterGenerator) Generate(ctx context.Context, pageURL, prevText, newText string, enabledTypes []string) ([]*entities.Insight, error) {
	if len(enabledTypes) == 0 {
		return nil, nil
	}

	// Always include overview as the first insight.
	allTypes := buildOrderedTypes(enabledTypes)

	prevText = truncate(prevText, maxTextLen)
	newText = truncate(newText, maxTextLen)

	var sections []string
	var jsonTemplate []string
	for i, t := range allTypes {
		desc, ok := insightTypeDescriptions[t]
		if !ok {
			desc = fmt.Sprintf("%s: analyse this aspect of the changes", t)
		}
		title := insightTitles[t]
		if title == "" {
			title = t
		}
		sections = append(sections, fmt.Sprintf("%d. %s", i+1, desc))
		jsonTemplate = append(jsonTemplate, fmt.Sprintf(`  {"insight_type": %q, "title": %q, "content": "<your analysis here>"}`, t, title))
	}

	prompt := fmt.Sprintf(`You are a senior competitive intelligence analyst. A monitoring system detected content changes on the following webpage:
%s

Your task: analyse the differences between the PREVIOUS and NEW page content below and produce exactly %d insight(s) in the order listed.

WRITING GUIDELINES
- Be specific and concrete — reference actual phrases, sections, or elements that changed.
- Each insight must be substantive: 2–4 paragraphs or clearly formatted numbered points.
- Do NOT pad with generic observations. Every sentence must add analytical value.
- Use plain text only in the "content" field — no markdown bold, no asterisks, no headers.
- Write in a confident, professional analyst voice.

INSIGHTS TO GENERATE (%d total, in this exact order):
%s

PREVIOUS PAGE CONTENT:
---
%s
---

NEW PAGE CONTENT:
+++
%s
+++

Return ONLY a valid JSON array with exactly %d object(s). No markdown fences, no explanation outside the array — pure JSON:
[
%s
]`,
		pageURL,
		len(allTypes),
		len(allTypes),
		strings.Join(sections, "\n\n"),
		prevText,
		newText,
		len(allTypes),
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

// buildOrderedTypes returns [overview, ...rest] deduplicating overview.
func buildOrderedTypes(enabled []string) []string {
	out := []string{"overview"}
	for _, t := range enabled {
		if t != "overview" {
			out = append(out, t)
		}
	}
	return out
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
