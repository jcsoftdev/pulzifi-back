package adminwiring

import (
	"context"

	adminservices "github.com/jcsoftdev/pulzifi-back/modules/admin/domain/services"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/templates"
)

// notifierAdapter implements adminservices.RegistrationNotifier by wrapping the
// email module's provider and composing the approval/rejection templates.
type notifierAdapter struct {
	provider emailservices.EmailProvider
}

// NewNotifierAdapter creates a RegistrationNotifier backed by the email module's provider.
func NewNotifierAdapter(provider emailservices.EmailProvider) adminservices.RegistrationNotifier {
	return &notifierAdapter{provider: provider}
}

func (a *notifierAdapter) SendApproval(ctx context.Context, toEmail, firstName, orgSubdomain, loginURL string) error {
	subject, html := templates.ApprovalNotification(firstName, orgSubdomain, loginURL)
	return a.provider.Send(ctx, toEmail, subject, html)
}

func (a *notifierAdapter) SendRejection(ctx context.Context, toEmail, firstName string) error {
	subject, html := templates.RejectionNotification(firstName)
	return a.provider.Send(ctx, toEmail, subject, html)
}
