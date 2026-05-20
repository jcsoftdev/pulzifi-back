package authwiring

import (
	"context"

	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/templates"
)

// notifierAdapter implements authservices.RegistrationNotifier by wrapping the
// email module's provider and composing the relevant templates.
type notifierAdapter struct {
	provider emailservices.EmailProvider
}

// NewNotifierAdapter creates a RegistrationNotifier backed by the email module's provider.
func NewNotifierAdapter(provider emailservices.EmailProvider) authservices.RegistrationNotifier {
	return &notifierAdapter{provider: provider}
}

func (a *notifierAdapter) SendRegistrationSubmitted(ctx context.Context, toEmail, firstName, orgName string) error {
	subject, html := templates.RegistrationSubmitted(firstName, orgName)
	return a.provider.Send(ctx, toEmail, subject, html)
}

func (a *notifierAdapter) SendPasswordReset(ctx context.Context, toEmail, firstName, resetURL string) error {
	subject, html := templates.PasswordReset(firstName, resetURL)
	return a.provider.Send(ctx, toEmail, subject, html)
}
