package mocks

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/insight/domain/services"
)

// compile-time interface check
var _ services.InsightQuotaReader = (*MockInsightQuotaReader)(nil)

// MockInsightQuotaReader is a hand-rolled test double for the AI insight
// quota port. Tests set the *Result fields or *Fn function hooks.
//
// When neither Fn nor Result is configured, Read returns
// {Unlimited:true} so handlers under test default to "no cap" — most tests
// only care about specific quota scenarios.
type MockInsightQuotaReader struct {
	// Read
	ReadResult services.InsightQuotaSnapshot
	ReadErr    error
	ReadFn     func(ctx context.Context, tenant string) (services.InsightQuotaSnapshot, error)

	// Increment
	IncrementResult services.InsightQuotaSnapshot
	IncrementErr    error
	IncrementFn     func(ctx context.Context, tenant string) (services.InsightQuotaSnapshot, error)

	// Call counters and last-args
	ReadCalls         int
	IncrementCalls    int
	LastReadTenant    string
	LastIncTenant     string
}

func (m *MockInsightQuotaReader) Read(ctx context.Context, tenant string) (services.InsightQuotaSnapshot, error) {
	m.ReadCalls++
	m.LastReadTenant = tenant
	if m.ReadFn != nil {
		return m.ReadFn(ctx, tenant)
	}
	if (m.ReadResult == services.InsightQuotaSnapshot{}) && m.ReadErr == nil {
		return services.InsightQuotaSnapshot{Unlimited: true}, nil
	}
	return m.ReadResult, m.ReadErr
}

func (m *MockInsightQuotaReader) Increment(ctx context.Context, tenant string) (services.InsightQuotaSnapshot, error) {
	m.IncrementCalls++
	m.LastIncTenant = tenant
	if m.IncrementFn != nil {
		return m.IncrementFn(ctx, tenant)
	}
	return m.IncrementResult, m.IncrementErr
}
