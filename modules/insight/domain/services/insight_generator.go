package services

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/entities"
)

// InsightGenerator defines the contract for generating strategic insights from page content changes.
// enabledTypes controls which insight types are produced (e.g. ["marketing","market_analysis"]).
type InsightGenerator interface {
	Generate(ctx context.Context, pageURL, prevText, newText string, enabledTypes []string) ([]*entities.Insight, error)
}
