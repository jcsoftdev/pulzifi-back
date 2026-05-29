package generateinsights

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/entities"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/repositories"
	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/services"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

var defaultInsightTypes = []string{"marketing", "market_analysis"}

// GenerateInsightsHandler orchestrates insight generation and persistence.
//
// Quota enforcement (Plan B model):
//   - Before invoking the (expensive) LLM, snapshot the tenant's AI insight
//     usage via InsightQuotaReader.Read; abort with ErrInsightQuotaExhausted
//     when used >= allowed and the plan is NOT unlimited.
//   - After each successful repo.Create, atomically increment the counter via
//     InsightQuotaReader.Increment. When the post-increment value overflows
//     the cap, break the loop — already-persisted insights remain; remaining
//     ones in the batch are dropped (acceptable: cap is hard, not metered).
type GenerateInsightsHandler struct {
	generator   services.InsightGenerator
	repo        repositories.InsightRepository
	quotaReader services.InsightQuotaReader
}

// NewGenerateInsightsHandler creates a new GenerateInsightsHandler.
// quotaReader is optional — when nil, the handler runs unguarded (legacy
// behaviour, used by tests that don't exercise quota).
func NewGenerateInsightsHandler(
	generator services.InsightGenerator,
	repo repositories.InsightRepository,
	quotaReader services.InsightQuotaReader,
) *GenerateInsightsHandler {
	return &GenerateInsightsHandler{
		generator:   generator,
		repo:        repo,
		quotaReader: quotaReader,
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

	// Pre-flight quota check — short-circuit BEFORE the LLM call when the
	// tenant has already exhausted their monthly allowance.
	if h.quotaExhausted(ctx, req.SchemaName) {
		return ErrInsightQuotaExhausted
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

	return h.persistInsights(ctx, req, insights)
}

// quotaExhausted reports whether the tenant has already used up its monthly AI
// insight allowance. Skipped (returns false) when no reader is wired (test /
// legacy paths) or no schema is provided. Quota read failures fail-open
// (return false) — blocking generation on a billing-adjacent read is riskier
// than letting it through.
func (h *GenerateInsightsHandler) quotaExhausted(ctx context.Context, schemaName string) bool {
	if h.quotaReader == nil || schemaName == "" {
		return false
	}
	snap, err := h.quotaReader.Read(ctx, schemaName)
	if err != nil {
		logger.Warn("insight quota read failed — proceeding without guard",
			zap.String("tenant", schemaName),
			zap.Error(err))
		return false
	}
	return !snap.Unlimited && snap.Used >= snap.Allowed
}

// persistInsights stores each generated insight and increments the quota
// counter, stopping early once the cap is hit mid-batch.
func (h *GenerateInsightsHandler) persistInsights(ctx context.Context, req *Request, insights []*entities.Insight) error {
	for _, insight := range insights {
		insight.ID = uuid.New()
		insight.PageID = req.PageID
		insight.CheckID = req.CheckID
		insight.CreatedAt = time.Now()
		if err := h.repo.Create(ctx, insight); err != nil {
			return fmt.Errorf("store insight %q: %w", insight.InsightType, err)
		}

		// Increment + early-stop on overflow. Increment failure must not
		// undo the successful Create — log and continue.
		if h.quotaReader != nil && req.SchemaName != "" {
			post, incErr := h.quotaReader.Increment(ctx, req.SchemaName)
			if incErr != nil {
				logger.Warn("insight quota increment failed — counter may be stale",
					zap.String("tenant", req.SchemaName),
					zap.Error(incErr))
				continue
			}
			if !post.Unlimited && post.Used >= post.Allowed {
				logger.Info("insight quota reached cap mid-batch — dropping remaining insights",
					zap.String("tenant", req.SchemaName),
					zap.Int("used", post.Used),
					zap.Int("allowed", post.Allowed))
				break
			}
		}
	}

	return nil
}

// ErrInsightQuotaExhausted re-exports the domain sentinel so callers in this
// package can errors.Is() against it without an extra import.
var ErrInsightQuotaExhausted = services.ErrInsightQuotaExhausted

// IsQuotaExhausted is a small helper for callers that want to differentiate
// quota-driven failures from genuine errors without importing the domain pkg.
func IsQuotaExhausted(err error) bool {
	return errors.Is(err, ErrInsightQuotaExhausted)
}
