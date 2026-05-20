package snapshotwiring

import (
	"context"

	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/templates"
	snapServices "github.com/jcsoftdev/pulzifi-back/modules/snapshot/domain/services"
)

// notificationEmailerAdapter implements snapshot's NotificationEmailer port by
// wrapping the email module's EmailProvider and alert email template.
type notificationEmailerAdapter struct {
	provider emailservices.EmailProvider
}

// NewNotificationEmailer builds a NotificationEmailer adapter backed by
// the given EmailProvider.
func NewNotificationEmailer(provider emailservices.EmailProvider) snapServices.NotificationEmailer {
	return &notificationEmailerAdapter{provider: provider}
}

func (a *notificationEmailerAdapter) SendAlertEmail(ctx context.Context, to, pageURL, changeType, dashboardURL string) error {
	subject, html := templates.AlertNotification(pageURL, changeType, dashboardURL)
	return a.provider.Send(ctx, to, subject, html)
}
