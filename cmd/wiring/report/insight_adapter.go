// Package reportwiring contains cross-module adapters for the report module.
// It bridges the report module to the insight module without a direct
// module-to-module import (which the architecture check forbids).
package reportwiring

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	insightpersistence "github.com/jcsoftdev/pulzifi-back/modules/insight/infrastructure/persistence"
	reportservices "github.com/jcsoftdev/pulzifi-back/modules/report/domain/services"
)

// InsightReaderAdapter adapts the insight module's repository to the report
// module's InsightReader port.
type InsightReaderAdapter struct {
	db     *sql.DB
	tenant string
}

// NewInsightReaderAdapter creates a per-tenant InsightReader backed by the
// insight module's PostgreSQL repository.
func NewInsightReaderAdapter(db *sql.DB, tenant string) *InsightReaderAdapter {
	return &InsightReaderAdapter{db: db, tenant: tenant}
}

// ListByPage returns the page's insights flattened into the report module's
// InsightSummary type.
func (a *InsightReaderAdapter) ListByPage(ctx context.Context, pageID uuid.UUID) ([]reportservices.InsightSummary, error) {
	repo := insightpersistence.NewInsightPostgresRepository(a.db, a.tenant)
	list, err := repo.ListByPageID(ctx, pageID)
	if err != nil {
		return nil, err
	}

	out := make([]reportservices.InsightSummary, 0, len(list))
	for _, in := range list {
		out = append(out, reportservices.InsightSummary{
			Type:      in.InsightType,
			Title:     in.Title,
			Content:   in.Content,
			CreatedAt: in.CreatedAt,
		})
	}
	return out, nil
}
