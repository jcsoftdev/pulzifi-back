package snapshotwiring

import (
	"context"
	"database/sql"

	generateinsights "github.com/jcsoftdev/pulzifi-back/modules/insight/application/generate_insights"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/services"
	insightAI "github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/ai"
	insightpersistence "github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/persistence"
	snapServices "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
	sharedAI "github.com/jcsoftdev/pulzifi-back/shared/ai"
	"github.com/jcsoftdev/pulzifi-back/shared/config"
)

// insightDispatcherAdapter implements snapshot's InsightDispatcher port by
// wrapping insight/application/generate_insights.GenerateInsightsHandler.
// The handler is created per-dispatch so the repository is tenant-scoped.
// The quotaReader is process-wide (tenant arrives in the Request) and shared
// across dispatches.
type insightDispatcherAdapter struct {
	db          *sql.DB
	generator   services.InsightGenerator
	quotaReader services.InsightQuotaReader
}

// NewInsightDispatcher builds an InsightDispatcher adapter. Returns nil when
// the OpenRouter API key is not configured (insights disabled).
func NewInsightDispatcher(db *sql.DB, cfg *config.Config) snapServices.InsightDispatcher {
	if cfg.OpenRouterAPIKey == "" {
		return nil
	}
	openRouterClient := sharedAI.NewOpenRouterClient(cfg.OpenRouterAPIKey, cfg.OpenRouterModel).WithBaseURL(cfg.AIBaseURL)
	generator := insightAI.NewOpenRouterGenerator(openRouterClient)
	quotaReader := insightpersistence.NewQuotaPostgresRepository(db)
	return &insightDispatcherAdapter{db: db, generator: generator, quotaReader: quotaReader}
}

func (a *insightDispatcherAdapter) Dispatch(ctx context.Context, req snapServices.InsightRequest) error {
	// Create a tenant-scoped insight repository per dispatch call.
	repo := insightpersistence.NewInsightPostgresRepository(a.db, req.SchemaName)
	handler := generateinsights.NewGenerateInsightsHandler(a.generator, repo, a.quotaReader)
	return handler.Handle(ctx, &generateinsights.Request{
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
