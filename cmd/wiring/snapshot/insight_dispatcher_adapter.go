package snapshotwiring

import (
	"context"
	"database/sql"

	generateinsights "github.com/jcsoftdev/pulzifi-back/modules/insight/application/generate_insights"
	insightAI "github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/ai"
	snapServices "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
	sharedAI "github.com/jcsoftdev/pulzifi-back/shared/ai"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
)

// insightDispatcherAdapter implements snapshot's InsightDispatcher port by
// wrapping insight/application/generate_insights.GenerateInsightsHandler.
type insightDispatcherAdapter struct {
	handler *generateinsights.GenerateInsightsHandler
}

// NewInsightDispatcher builds an InsightDispatcher adapter. Returns nil when
// the OpenRouter API key is not configured (insights disabled).
func NewInsightDispatcher(db *sql.DB, cfg *config.Config) snapServices.InsightDispatcher {
	if cfg.OpenRouterAPIKey == "" {
		return nil
	}
	openRouterClient := sharedAI.NewOpenRouterClient(cfg.OpenRouterAPIKey, cfg.OpenRouterModel)
	generator := insightAI.NewOpenRouterGenerator(openRouterClient)
	handler := generateinsights.NewGenerateInsightsHandler(generator, db)
	return &insightDispatcherAdapter{handler: handler}
}

func (a *insightDispatcherAdapter) Dispatch(ctx context.Context, req snapServices.InsightRequest) error {
	return a.handler.Handle(ctx, &generateinsights.Request{
		PageID:              req.PageID,
		CheckID:             req.CheckID,
		PageURL:             req.PageURL,
		PrevText:            req.PrevText,
		NewText:             req.NewText,
		DiffText:            req.DiffText,
		SchemaName:          req.SchemaName,
		EnabledInsightTypes: req.EnabledInsightTypes,
	})
}
