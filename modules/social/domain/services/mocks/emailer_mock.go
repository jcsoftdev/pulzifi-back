package mocks

import (
	"context"

	"github.com/google/uuid"
)

// MockSocialNotificationEmailer is a hand-rolled mock for services.SocialNotificationEmailer.
type MockSocialNotificationEmailer struct {
	SendChangeEmailErr   error
	SendChangeEmailFn    func(ctx context.Context, workspaceID uuid.UUID, handle, platform, summary, dashboardURL string) error
	SendChangeEmailCalls int
}

func (m *MockSocialNotificationEmailer) SendChangeEmail(ctx context.Context, workspaceID uuid.UUID, handle, platform, summary, dashboardURL string) error {
	m.SendChangeEmailCalls++
	if m.SendChangeEmailFn != nil {
		return m.SendChangeEmailFn(ctx, workspaceID, handle, platform, summary, dashboardURL)
	}
	return m.SendChangeEmailErr
}
