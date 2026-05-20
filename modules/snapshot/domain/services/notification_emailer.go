package services

import "context"

// NotificationEmailer is the port for sending change-detection alert emails.
// The concrete adapter wraps email/domain/services.EmailProvider and
// email/infrastructure/templates.
type NotificationEmailer interface {
	// SendAlertEmail sends a change-detection alert email to a recipient.
	SendAlertEmail(ctx context.Context, to, pageURL, changeType, dashboardURL string) error
}
