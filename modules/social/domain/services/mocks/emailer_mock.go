package mocks

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// MockSocialNotificationEmailer is a hand-rolled mock for services.SocialNotificationEmailer.
// SendChangeEmail is safe for concurrent use because run_check calls it from a goroutine.
type MockSocialNotificationEmailer struct {
	mu                   sync.Mutex
	SendChangeEmailErr   error
	SendChangeEmailFn    func(ctx context.Context, workspaceID uuid.UUID, handle, platform, summary, dashboardURL string) error
	SendChangeEmailCalls int
}

func (m *MockSocialNotificationEmailer) SendChangeEmail(ctx context.Context, workspaceID uuid.UUID, handle, platform, summary, dashboardURL string) error {
	m.mu.Lock()
	m.SendChangeEmailCalls++
	err := m.SendChangeEmailErr
	fn := m.SendChangeEmailFn
	m.mu.Unlock()
	if fn != nil {
		return fn(ctx, workspaceID, handle, platform, summary, dashboardURL)
	}
	return err
}

// Calls returns the send count under lock (use in tests after goroutines settle).
func (m *MockSocialNotificationEmailer) Calls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.SendChangeEmailCalls
}
