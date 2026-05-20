package generateinsights

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/services"
)

var defaultInsightTypes = []string{"marketing", "market_analysis"}

// GenerateInsightsHandler orchestrates insight generation and persistence.
type GenerateInsightsHandler struct {
	generator services.InsightGenerator
	repo      repositories.InsightRepository
}

// NewGenerateInsightsHandler creates a new GenerateInsightsHandler.
func NewGenerateInsightsHandler(generator services.InsightGenerator, repo repositories.InsightRepository) *GenerateInsightsHandler {
	return &GenerateInsightsHandler{
		generator: generator,
		repo:      repo,
	}
}

// Handle generates insights for a detected page change and stores them.
// When DiffText is provided (from content block diffing), it uses the more
// token-efficient GenerateFromDiff path. Falls back to full-text Generate.
func (h *GenerateInsightsHandler) Handle(ctx context.Context, req *Request) error {
	enabledTypes := req.EnabledInsightTypes
	if len(enabledTypes) == 0 {
		enabledTypes = defaultInsightTypes
	}

	var insights []*entities.Insight
	var err error

	if req.DiffText != "" {
		insights, err = h.generator.GenerateFromDiff(ctx, req.PageURL, req.DiffText, enabledTypes)
	} else {
		insights, err = h.generator.Generate(ctx, req.PageURL, req.PrevText, req.NewText, enabledTypes)
	}
	if err != nil {
		return fmt.Errorf("generate insights: %w", err)
	}
	if len(insights) == 0 {
		return nil
	}

	for _, insight := range insights {
		insight.ID = uuid.New()
		insight.PageID = req.PageID
		insight.CheckID = req.CheckID
		insight.CreatedAt = time.Now()
		if err := h.repo.Create(ctx, insight); err != nil {
			return fmt.Errorf("store insight %q: %w", insight.InsightType, err)
		}
	}

	return nil
}
