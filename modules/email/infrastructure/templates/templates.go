package templates

import "fmt"

func wrap(title, body string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><title>%s</title></head>
<body style="font-family:Arial,sans-serif;max-width:600px;margin:0 auto;padding:20px;color:#333;">
%s
<hr style="border:none;border-top:1px solid #eee;margin-top:30px;">
<p style="font-size:12px;color:#999;">This email was sent by Pulzifi. If you did not expect this, please ignore it.</p>
</body>
</html>`, title, body)
}

// ApprovalNotification generates the approval notification email.
func ApprovalNotification(firstName, subdomain, loginURL string) (subject, html string) {
	subject = "Your Pulzifi account has been approved"
	html = wrap(subject, fmt.Sprintf(`
<h2>Welcome to Pulzifi, %s!</h2>
<p>Your account has been approved. Your organization is ready at <strong>%s</strong>.</p>
<p><a href="%s" style="display:inline-block;background:#4F46E5;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;">Log in to your dashboard</a></p>
`, firstName, subdomain, loginURL))
	return
}

// RejectionNotification generates the rejection notification email.
func RejectionNotification(firstName string) (subject, html string) {
	subject = "Your Pulzifi registration update"
	html = wrap(subject, fmt.Sprintf(`
<h2>Hi %s,</h2>
<p>We've reviewed your registration request and unfortunately we're unable to approve your account at this time.</p>
<p>If you believe this is an error, please contact our support team.</p>
`, firstName))
	return
}

// TeamInvite generates the team invitation email.
func TeamInvite(inviterName, orgName, loginURL string) (subject, html string) {
	subject = fmt.Sprintf("You've been invited to %s on Pulzifi", orgName)
	html = wrap(subject, fmt.Sprintf(`
<h2>You've been invited!</h2>
<p><strong>%s</strong> has invited you to join <strong>%s</strong> on Pulzifi.</p>
<p><a href="%s" style="display:inline-block;background:#4F46E5;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;">Accept Invitation</a></p>
`, inviterName, orgName, loginURL))
	return
}

// PasswordReset generates the password reset email.
func PasswordReset(firstName, resetURL string) (subject, html string) {
	subject = "Reset your Pulzifi password"
	html = wrap(subject, fmt.Sprintf(`
<h2>Hi %s,</h2>
<p>We received a request to reset your password. Click the button below to choose a new one. This link expires in 1 hour.</p>
<p><a href="%s" style="display:inline-block;background:#4F46E5;color:#fff;padding:12px 24px;border-radius:6px;text-decoration:none;">Reset Password</a></p>
<p>If you didn't request this, you can safely ignore this email.</p>
`, firstName, resetURL))
	return
}
