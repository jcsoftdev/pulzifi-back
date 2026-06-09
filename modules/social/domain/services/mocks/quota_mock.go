package mocks

import (
	"context"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
)

// MockCheckQuota is a hand-rolled mock for services.CheckQuota.
type MockCheckQuota struct {
	ConsumeResult services.QuotaResult
	ConsumeErr    error
	CompensateErr error

	ConsumeFn    func(ctx context.Context, date time.Time, limit int) (services.QuotaResult, error)
	CompensateFn func(ctx context.Context, date time.Time) error

	ConsumeCalls    int
	CompensateCalls int
}

func (m *MockCheckQuota) Consume(ctx context.Context, date time.Time, limit int) (services.QuotaResult, error) {
	m.ConsumeCalls++
	if m.ConsumeFn != nil {
		return m.ConsumeFn(ctx, date, limit)
	}
	return m.ConsumeResult, m.ConsumeErr
}

func (m *MockCheckQuota) Compensate(ctx context.Context, date time.Time) error {
	m.CompensateCalls++
	if m.CompensateFn != nil {
		return m.CompensateFn(ctx, date)
	}
	return m.CompensateErr
}
