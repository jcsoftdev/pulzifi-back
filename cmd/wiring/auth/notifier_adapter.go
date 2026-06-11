package authwiring

import (
	"context"
	"strings"

	authservices "github.com/jcsoftdev/pulzifi-back/modules/auth/domain/services"
	emailservices "github.com/jcsoftdev/pulzifi-back/modules/email/domain/services"
	"github.com/jcsoftdev/pulzifi-back/modules/email/infrastructure/templates"
)

// notifierAdapter implements authservices.RegistrationNotifier by wrapping the
// email module's provider and composing the relevant templates.
type notifierAdapter struct {
	provider  emailservices.EmailProvider
	loginURL  string
	trialDays int
}

// NewNotifierAdapter creates a RegistrationNotifier backed by the email module's provider.
// loginURL is the apex frontend URL used in the welcome CTA; trialDays personalizes
// the trial-length copy.
func NewNotifierAdapter(provider emailservices.EmailProvider, loginURL string, trialDays int) authservices.RegistrationNotifier {
	return &notifierAdapter{provider: provider, loginURL: loginURL, trialDays: trialDays}
}

func (a *notifierAdapter) SendRegistrationSubmitted(ctx context.Context, toEmail, firstName, orgName, planCode string) error {
	// Self-serve signup is auto-approved — send the day-0 welcome email, not
	// the legacy "under review" notice. A paid plan selected on the pricing
	// page gets checkout-oriented copy instead of trial copy.
	var subject, html string
	if planCode != "" && planCode != "trial" && planCode != "free" {
		planName := strings.ToUpper(planCode[:1]) + planCode[1:]
		subject, html = templates.WelcomePlanSelected(firstName, orgName, planName, a.loginURL)
	} else {
		subject, html = templates.WelcomeTrial(firstName, orgName, a.loginURL, a.trialDays)
	}
	return a.provider.Send(ctx, toEmail, subject, html)
}

func (a *notifierAdapter) SendPasswordReset(ctx context.Context, toEmail, firstName, resetURL string) error {
	subject, html := templates.PasswordReset(firstName, resetURL)
	return a.provider.Send(ctx, toEmail, subject, html)
}
