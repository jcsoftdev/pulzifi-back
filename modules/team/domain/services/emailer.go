package services

import "context"

// Emailer abstracts sending invitation emails.
// The concrete implementation is injected at startup from the email module.
type Emailer interface {
	Send(ctx context.Context, to, subject, html string) error
}

// InviteEmailBuilder builds the subject and HTML body for a team invitation email.
// The concrete implementation uses the email module's templates.
type InviteEmailBuilder interface {
	TeamInvite(inviterEmail, subdomain, loginURL string) (subject, html string)
}
