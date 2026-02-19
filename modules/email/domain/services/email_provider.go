package services

import "context"

// EmailProvider defines the interface for sending emails.
// Implementations can use Resend, SES, SMTP, etc.
type EmailProvider interface {
	Send(ctx context.Context, to, subject, htmlBody string) error
}
