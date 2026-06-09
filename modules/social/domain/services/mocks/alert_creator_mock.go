package mocks

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/social/domain/services"
)

// MockAlertCreator is a hand-rolled mock for services.AlertCreator.
type MockAlertCreator struct {
	CreateAlertErr error
	CreateAlertFn  func(ctx context.Context, payload services.AlertPayload) error

	CreateAlertCalls    int
	LastAlertPayload    services.AlertPayload
}

func (m *MockAlertCreator) CreateAlert(ctx context.Context, payload services.AlertPayload) error {
	m.CreateAlertCalls++
	m.LastAlertPayload = payload
	if m.CreateAlertFn != nil {
		return m.CreateAlertFn(ctx, payload)
	}
	return m.CreateAlertErr
}
