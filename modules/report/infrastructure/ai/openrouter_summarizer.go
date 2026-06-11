package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/jcsoftdev/pulzifi-back/modules/report/domain/services"
	sharedAI "github.com/jcsoftdev/pulzifi-back/shared/ai"
)

// maxInsights caps how many insights are fed to the LLM to keep the prompt
// within a reasonable token budget. Insights arrive newest-first.
const maxInsights = 50

// maxInsightContentLen truncates each insight body so a few verbose insights
// cannot blow the prompt size.
const maxInsightContentLen = 1500

// OpenRouterSummarizer implements ReportSummarizer using the OpenRouter API.
type OpenRouterSummarizer struct {
	client *sharedAI.OpenRouterClient
}

// NewOpenRouterSummarizer creates a new OpenRouterSummarizer.
func NewOpenRouterSummarizer(client *sharedAI.OpenRouterClient) *OpenRouterSummarizer {
	return &OpenRouterSummarizer{client: client}
}

// Summarize produces a plain-text monthly report from the page's AI insights:
// a short overview paragraph, plus bullet-point recommended actions when warranted.
func (s *OpenRouterSummarizer) Summarize(ctx context.Context, pageTitle string, insights []services.InsightSummary) (string, error) {
	// Insights arrive newest-first (ListByPageID orders by created_at DESC), so
	// capping to the head keeps the most recent insights.
	if len(insights) > maxInsights {
		insights = insights[:maxInsights]
	}

	var b strings.Builder
	for i, in := range insights {
		title := in.Title
		if title == "" {
			title = in.Type
		}
		fmt.Fprintf(&b, "%d. [%s] %s\n%s\n\n", i+1, in.Type, title, truncate(in.Content, maxInsightContentLen))
	}

	prompt := fmt.Sprintf(`You are a competitive intelligence analyst. Below are the AI insights generated over time for the monitored page %q.

Write a concise MONTHLY REPORT that summarizes these insights for a busy stakeholder.

FORMAT RULES (follow exactly):
- Plain text only. No markdown headers, no bold, no asterisks for emphasis.
- Start with a short paragraph (2-4 sentences) covering the overall situation and the most important changes.
- THEN, only if the insights clearly warrant action, add a line that reads exactly "Recommended actions:" followed by bullet points, one per line, each starting with "- ". Keep each action concise and concrete.
- If no actions are warranted, omit the actions section entirely.
- Do not invent information that is not present in the insights below.

INSIGHTS:
%s`, pageTitle, b.String())

	messages := []sharedAI.Message{{Role: "user", Content: prompt}}

	out, err := s.client.Complete(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("report summarizer: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
