package errors

import "fmt"

// InvitationError represents invitation-related errors
type InvitationError struct {
	Code    string
	Message string
}

func (e InvitationError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewInvitationError(code, message string) InvitationError {
	return InvitationError{Code: code, Message: message}
}

var (
	ErrCannotInviteEmail        = InvitationError{Code: "INVITATION_CANNOT_INVITE_EMAIL", Message: "cannot invite this email"}
	ErrInvitationNotFound       = InvitationError{Code: "INVITATION_NOT_FOUND", Message: "invitation not found or expired"}
	ErrDailyCapExceeded         = InvitationError{Code: "INVITATION_DAILY_CAP_EXCEEDED", Message: "daily invitation cap exceeded"}
	ErrInvitationAlreadyDecided = InvitationError{Code: "INVITATION_ALREADY_DECIDED", Message: "invitation already accepted or revoked"}
	ErrSubdomainTaken           = InvitationError{Code: "INVITATION_SUBDOMAIN_TAKEN", Message: "subdomain already taken"}
	ErrInvalidSubdomain         = InvitationError{Code: "INVITATION_INVALID_SUBDOMAIN", Message: "invalid subdomain"}
	ErrSchemaProvisioning       = InvitationError{Code: "INVITATION_SCHEMA_PROVISIONING_FAILED", Message: "schema provisioning failed"}
	ErrSessionIssueFailed       = InvitationError{Code: "INVITATION_SESSION_ISSUE_FAILED", Message: "session issuance failed"}
)
