package mocks

import (
	"context"

	"github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
)

// MockAlertCreator is a hand-rolled mock for services.AlertCreator.
type MockAlertCreator struct {
	CreateErr error

	CreateFn func(ctx context.Context, input services.AlertInput) error

	CreateCalls int
}

func (m *MockAlertCreator) Create(ctx context.Context, input services.AlertInput) error {
	m.CreateCalls++
	if m.CreateFn != nil {
		return m.CreateFn(ctx, input)
	}
	return m.CreateErr
}
