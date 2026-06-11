package mocks

import "context"

// MockNotifier is a mock of services.RegistrationNotifier.
type MockNotifier struct {
	SendRegistrationSubmittedErr error
	SendPasswordResetErr         error
}

func (m *MockNotifier) SendRegistrationSubmitted(_ context.Context, _, _, _, _ string) error {
	return m.SendRegistrationSubmittedErr
}

func (m *MockNotifier) SendPasswordReset(_ context.Context, _, _, _ string) error {
	return m.SendPasswordResetErr
}
