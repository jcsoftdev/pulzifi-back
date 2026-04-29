package services

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/entities"
)

// InsightGenerator defines the contract for generating strategic insights from page content changes.
// enabledTypes controls which insight types are produced (e.g. ["marketing","market_analysis"]).
type InsightGenerator interface {
	Generate(ctx context.Context, pageURL, prevText, newText string, enabledTypes []string) ([]*entities.Insight, error)
	// GenerateFromDiff generates insights from a compact structural diff text
	// instead of full previous/new text. Uses significantly fewer tokens.
	GenerateFromDiff(ctx context.Context, pageURL, diffText string, enabledTypes []string) ([]*entities.Insight, error)
}
