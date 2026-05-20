package services

import "context"

// RegistrationNotifier is auth's port for sending registration-related emails.
// Implemented by authwiring in cmd/wiring/auth/.
type RegistrationNotifier interface {
	// SendRegistrationSubmitted notifies the user that their registration is
	// under review.
	SendRegistrationSubmitted(ctx context.Context, toEmail, firstName, orgName string) error

	// SendPasswordReset sends a password-reset link to the user.
	SendPasswordReset(ctx context.Context, toEmail, firstName, resetURL string) error
}
