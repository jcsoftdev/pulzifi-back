package services

import "context"

// RegistrationNotifier is auth's port for sending registration-related emails.
// Implemented by authwiring in cmd/wiring/auth/.
type RegistrationNotifier interface {
	// SendRegistrationSubmitted sends the day-0 welcome email. planCode is the
	// paid plan selected on the pricing page ("" for the free trial signup) and
	// switches the copy between trial-started and plan-checkout-pending.
	SendRegistrationSubmitted(ctx context.Context, toEmail, firstName, orgName, planCode string) error

	// SendPasswordReset sends a password-reset link to the user.
	SendPasswordReset(ctx context.Context, toEmail, firstName, resetURL string) error
}
