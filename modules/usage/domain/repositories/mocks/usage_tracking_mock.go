package mocks

import (
	"context"
	"time"

	"github.com/jcsoftdev/pulzifi-back/modules/usage/domain/entities"
)

// MockUsageTrackingRepository is a hand-rolled mock of repositories.UsageTrackingRepository.
type MockUsageTrackingRepository struct {
	FindCurrentResult     *entities.UsageTracking
	FindCurrentErr        error
	InsertErr             error
	LatestPeriodEndResult time.Time
	LatestPeriodEndErr    error
	SyncChecksAllowedErr  error
	TotalChecks           int
	SuccessChecks         int
	FailedChecks          int
	CountChecksErr        error
	TotalPages            int
	CountPagesErr         error
	TotalWorkspaces       int
	CountWorkspacesErr    error
	TotalAlerts           int
	CountAlertsErr        error
}

func (m *MockUsageTrackingRepository) FindCurrent(_ context.Context, _ time.Time) (*entities.UsageTracking, error) {
	return m.FindCurrentResult, m.FindCurrentErr
}

func (m *MockUsageTrackingRepository) Insert(_ context.Context, _ *entities.UsageTracking) error {
	return m.InsertErr
}

func (m *MockUsageTrackingRepository) LatestPeriodEnd(_ context.Context) (time.Time, error) {
	return m.LatestPeriodEndResult, m.LatestPeriodEndErr
}

func (m *MockUsageTrackingRepository) SyncChecksAllowed(_ context.Context, _ int) error {
	return m.SyncChecksAllowedErr
}

func (m *MockUsageTrackingRepository) CountChecks(_ context.Context) (int, int, int, error) {
	return m.TotalChecks, m.SuccessChecks, m.FailedChecks, m.CountChecksErr
}

func (m *MockUsageTrackingRepository) CountPages(_ context.Context) (int, error) {
	return m.TotalPages, m.CountPagesErr
}

func (m *MockUsageTrackingRepository) CountWorkspaces(_ context.Context) (int, error) {
	return m.TotalWorkspaces, m.CountWorkspacesErr
}

func (m *MockUsageTrackingRepository) CountAlerts(_ context.Context) (int, error) {
	return m.TotalAlerts, m.CountAlertsErr
}
