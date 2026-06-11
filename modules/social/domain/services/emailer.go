package services

import (
	"context"

	"github.com/google/uuid"
)

// SocialNotificationEmailer sends a transactional email when a social change is
// detected. The concrete adapter (cmd/wiring/social/emailer_adapter.go) resolves
// recipient emails for the workspace and renders the HTML body.
type SocialNotificationEmailer interface {
	SendChangeEmail(ctx context.Context, workspaceID uuid.UUID, handle, platform, summary, dashboardURL string) error
}
