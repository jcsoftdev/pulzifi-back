package generateinsights

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/services"
	insightPersistence "github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/persistence"
)

var defaultInsightTypes = []string{"marketing", "market_analysis"}

// GenerateInsightsHandler orchestrates insight generation and persistence.
type GenerateInsightsHandler struct {
	generator services.InsightGenerator
	db        *sql.DB
}

// NewGenerateInsightsHandler creates a new GenerateInsightsHandler.
func NewGenerateInsightsHandler(generator services.InsightGenerator, db *sql.DB) *GenerateInsightsHandler {
	return &GenerateInsightsHandler{
		generator: generator,
		db:        db,
	}
}

// Handle generates insights for a detected page change and stores them.
func (h *GenerateInsightsHandler) Handle(ctx context.Context, req *Request) error {
	enabledTypes := req.EnabledInsightTypes
	if len(enabledTypes) == 0 {
		enabledTypes = defaultInsightTypes
	}

	insights, err := h.generator.Generate(ctx, req.PageURL, req.PrevText, req.NewText, enabledTypes)
	if err != nil {
		return fmt.Errorf("generate insights: %w", err)
	}
	if len(insights) == 0 {
		return nil
	}

	repo := insightPersistence.NewInsightPostgresRepository(h.db, req.SchemaName)
	for _, insight := range insights {
		insight.ID = uuid.New()
		insight.PageID = req.PageID
		insight.CheckID = req.CheckID
		insight.CreatedAt = time.Now()
		if err := repo.Create(ctx, insight); err != nil {
			return fmt.Errorf("store insight %q: %w", insight.InsightType, err)
		}
	}

	return nil
}
