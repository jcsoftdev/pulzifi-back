package mocks

import (
	"context"

	"github.com/google/uuid"
)

// MockUsageRepository is a hand-rolled test double for the orchestrator.UsageRepository interface.
type MockUsageRepository struct {
	HasQuotaResult bool
	HasQuotaErr    error
	LogUsageErr    error
	HasQuotaFn     func(ctx context.Context) (bool, error)
	LogUsageFn     func(ctx context.Context, pageID, checkID uuid.UUID) error

	HasQuotaCalls  int
	LogUsageCalls  int
}

func (m *MockUsageRepository) HasQuota(ctx context.Context) (bool, error) {
	m.HasQuotaCalls++
	if m.HasQuotaFn != nil {
		return m.HasQuotaFn(ctx)
	}
	return m.HasQuotaResult, m.HasQuotaErr
}

func (m *MockUsageRepository) LogUsage(ctx context.Context, pageID, checkID uuid.UUID) error {
	m.LogUsageCalls++
	if m.LogUsageFn != nil {
		return m.LogUsageFn(ctx, pageID, checkID)
	}
	return m.LogUsageErr
}
