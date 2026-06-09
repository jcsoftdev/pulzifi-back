package mocks

import "context"

// MockPlanLimits is a hand-rolled mock for services.PlanLimits.
type MockPlanLimits struct {
	GetProfilesLimitResult int
	GetProfilesLimitErr    error
	GetChecksPerDayResult  int
	GetChecksPerDayErr     error
	GetChecksUsedResult    int
	GetChecksUsedErr       error

	GetProfilesLimitFn   func(ctx context.Context, workspaceID string) (int, error)
	GetChecksPerDayFn    func(ctx context.Context, workspaceID string) (int, error)
	GetChecksUsedTodayFn func(ctx context.Context) (int, error)

	GetProfilesLimitCalls int
	GetChecksPerDayCalls  int
	GetChecksUsedCalls    int
}

func (m *MockPlanLimits) GetProfilesLimit(ctx context.Context, workspaceID string) (int, error) {
	m.GetProfilesLimitCalls++
	if m.GetProfilesLimitFn != nil {
		return m.GetProfilesLimitFn(ctx, workspaceID)
	}
	return m.GetProfilesLimitResult, m.GetProfilesLimitErr
}

func (m *MockPlanLimits) GetChecksPerDay(ctx context.Context, workspaceID string) (int, error) {
	m.GetChecksPerDayCalls++
	if m.GetChecksPerDayFn != nil {
		return m.GetChecksPerDayFn(ctx, workspaceID)
	}
	return m.GetChecksPerDayResult, m.GetChecksPerDayErr
}

func (m *MockPlanLimits) GetChecksUsedToday(ctx context.Context) (int, error) {
	m.GetChecksUsedCalls++
	if m.GetChecksUsedTodayFn != nil {
		return m.GetChecksUsedTodayFn(ctx)
	}
	return m.GetChecksUsedResult, m.GetChecksUsedErr
}
